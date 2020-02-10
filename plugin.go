/*
Drone plugin to upload one or more packages to Bintray.
See README.md for usage.
Author: Archit Sharma December 2019 (Github arcolife)
Previous: David Tootill November 2015 (GitHub tooda02)
*/

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arcolife/jfrog-client-go/bintray"
	"github.com/arcolife/jfrog-client-go/bintray/auth"
	"github.com/arcolife/jfrog-client-go/bintray/services"
	"github.com/arcolife/jfrog-client-go/bintray/services/packages"
	"github.com/arcolife/jfrog-client-go/bintray/services/repositories"
	"github.com/arcolife/jfrog-client-go/bintray/services/versions"
	"github.com/pkg/errors"
	// "github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type (
	// Package - struct fields must be public in order for unmarshal to
	// correctly populate the data.
	Package struct {
		Package string                `yaml:"package"`
		Config  packages.Params       `yaml:"config"`
		Upload  services.UploadParams `yaml:"upload"`
	}

	// Repo - struct fields must be public in order for unmarshal to
	// correctly populate the data.
	Repo struct {
		Name     string              `yaml:"name"`
		Subject  string              `yaml:"subject"`
		Config   repositories.Config `yaml:"config"`
		Packages []Package           `yaml:"packages"`
	}

	// Bintray - struct fields must be public in order for unmarshal to
	// correctly populate the data.
	Bintray struct {
		Repos []Repo `yaml:"repos"`
	}

	// BintrayConfig -> entrypoint for all configs
	BintrayConfig struct {
		Threads       int  `yaml:"threads"`
		Cleanup       bool `yaml:"cleanup"`
		Bintray       `yaml:"bintray"`
		Username      string `yaml:"username"`
		APIKey        string `yaml:"api_key"`
		GPGPassphrase string `yaml:"gpg_passphrase"`
	}

	Plugin struct {
		BintrayConfig      `yaml:"bintray_config"`
		ServicesManager    *bintray.ServicesManager
		BintrayConfigPath  string   `yaml:"bintray_cfg_path"`
		Version            string   `yaml:"verison"`
		Cleanup            bool     `yaml:"cleanup"`
		UploadPackage      bool     `yaml:"uploadpackage"`
		SignPackageVersion bool     `yaml:"signpackageversion"`
		PublishPackage     bool     `yaml:"publishpackage"`
		CalcMetadata       bool     `yaml:"calcmetadata"`
		ShowPackage        bool     `yaml:"showpackage"`
		Envs               []string `yaml:"envs"`
	}
)

const defaultLicense = "Apache 2.0"

func (p Plugin) Exec() error {
	// var (
	// 	files []string
	// )
	// fmt.Printf("Drone Bintray Plugin built from %s\n", p.Version)
	cSig := make(chan os.Signal)
	signal.Notify(cSig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSig
		cleanup()
		os.Exit(1)
	}()

	if p.BintrayConfigPath != "" {
		p.BintrayConfig.ReadConfig(p.BintrayConfigPath)
	}

	p.ServicesManager = p.BintrayConfig.InitConfig()

	if p.BintrayConfig.Cleanup || p.Cleanup {
		fmt.Printf("\nCleaning up")
		p.BintrayConfig.cleanup(p.ServicesManager)
		p.BintrayConfig.checkDetails(p.ServicesManager)
	}

	for _, repo := range p.BintrayConfig.Bintray.Repos {
		for _, pack := range repo.Packages {
			uploaded := 0
			if p.UploadPackage {
				uploaded, _, _ = p.BintrayConfig.uploadPackage(p.ServicesManager, &repo, &pack)
			}
			if uploaded > 0 || p.SignPackageVersion {
				p.BintrayConfig.signPackageVersion(p.ServicesManager, &repo, &pack)
			}
			if uploaded > 0 || (pack.Upload.Publish && p.PublishPackage) {
				p.BintrayConfig.publishPackage(p.ServicesManager, &repo, &pack)
			}
			if uploaded > 0 || p.CalcMetadata {
				p.BintrayConfig.calcMetadata(p.ServicesManager, &repo, &pack)
			}
			if uploaded > 0 || p.ShowPackage {
				p.BintrayConfig.showPackage(p.ServicesManager, &repo, &pack)
			}
		}
	}
	return nil
}

// InitConfig for initializing service manager for bintray functionalities
func (config *BintrayConfig) InitConfig() *bintray.ServicesManager {

	btDetails := auth.NewBintrayDetails()
	btDetails.SetApiUrl("https://api.bintray.com/")
	fmt.Printf("API url: [%s]\n", btDetails.GetApiUrl())
	btDetails.SetUser(config.Username)
	btDetails.SetKey(config.APIKey)
	btDetails.SetDefPackageLicense(defaultLicense)

	serviceConfig := bintray.NewConfigBuilder().
		SetBintrayDetails(btDetails).
		SetDryRun(false).
		SetThreads(config.Threads).
		Build()

	btManager, err := bintray.New(serviceConfig)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}

	return btManager
}

func (config *BintrayConfig) deleteRepo(btManager *bintray.ServicesManager, repo *Repo) error {
	fmt.Printf("\nDeleting Repo.. [%s]\n", repo.Name)
	repoPath := repositories.Path{Subject: repo.Subject, Repo: repo.Name}
	return errors.Wrap(btManager.ExecDeleteRepoRest(&repoPath), "Repo non-existent")
}

func (config *BintrayConfig) deletePackage(btManager *bintray.ServicesManager, repo *Repo, pack *Package) error {
	pkg := fmt.Sprintf("%s/%s/%s", repo.Subject, repo.Name, pack.Package)
	fmt.Printf("\nDeleting Package.. [%s]\n", pkg)
	packagePath, _ := packages.CreatePath(pkg)
	return errors.Wrap(btManager.DeletePackage(packagePath), "Package non-existent")
}

func (config *BintrayConfig) createRepo(btManager *bintray.ServicesManager, repo *Repo) error {
	var err error
	var existsOk bool

	repoPath := repositories.Path{Subject: repo.Subject, Repo: repo.Name}
	existsOk, err = btManager.CreateReposIfNeeded(&repoPath, &repo.Config, repo.Config.RepoConfigFilePath)
	if existsOk == true && err == nil {
		fmt.Println("Success")
	}
	return errors.Wrap(err, "RepoConfigFilePath non-existent")
}

func (config *BintrayConfig) createPackage(btManager *bintray.ServicesManager, repo *Repo, pack *Package) error {
	pkg := fmt.Sprintf("%s/%s/%s", repo.Subject, repo.Name, pack.Package)
	pack.Config.Path, _ = packages.CreatePath(pkg)
	return btManager.CreatePackage(&pack.Config)
}

func (config *BintrayConfig) publishPackage(btManager *bintray.ServicesManager, repo *Repo, pack *Package) error {
	pkg := fmt.Sprintf("%s/%s/%s", repo.Subject, repo.Name, pack.Package)
	fmt.Printf("\nPublishing GPG Signatures.. [%s]", pkg)
	versionPathString := fmt.Sprintf("%s/%s", pkg, pack.Upload.Version)
	versionPath, _ := versions.CreatePath(versionPathString)
	err := btManager.PublishVersion(versionPath)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return err
}

func (config *BintrayConfig) uploadPackage(btManager *bintray.ServicesManager, repo *Repo, pack *Package) (totalUploaded, totalFailed int, err error) {
	pkg := fmt.Sprintf("%s/%s/%s", repo.Subject, repo.Name, pack.Package)
	fmt.Printf("\nUploading Files to Package [%s] with Publish: [%t]\n", pkg, pack.Upload.Publish)
	versionPath := fmt.Sprintf("%s/%s", pkg, pack.Upload.Version)
	pack.Upload.Path, _ = versions.CreatePath(versionPath)
	PrettyPrint(&pack, "Package")
	totalUploaded, totalFailed, err = btManager.UploadFiles(&pack.Upload)
	fmt.Println("UPLOADED", totalUploaded)
	fmt.Println("FAILED: ", totalFailed)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return totalUploaded, totalFailed, err
}

func (config *BintrayConfig) signPackageVersion(btManager *bintray.ServicesManager, repo *Repo, pack *Package) error {
	pkg := fmt.Sprintf("%s/%s/%s", repo.Subject, repo.Name, pack.Package)
	fmt.Printf("\nSigning versioned Package files.. [%s]", pkg)
	versionPath := fmt.Sprintf("%s/%s", pkg, pack.Upload.Version)
	path, _ := versions.CreatePath(versionPath)
	err := btManager.GpgSignVersion(path, config.GPGPassphrase)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (config *BintrayConfig) calcMetadata(btManager *bintray.ServicesManager, repo *Repo, pack *Package) bool {
	pkg := fmt.Sprintf("%s/%s/%s", repo.Subject, repo.Name, pack.Package)
	fmt.Printf("\nScheduling metadata calculation.. [%s]\n", pkg)
	versionPath := fmt.Sprintf("%s/%s", pkg, pack.Upload.Version)
	path, _ := versions.CreatePath(versionPath)
	scheduledOk, err := btManager.CalcMetadata(path)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return scheduledOk
}

func (config *BintrayConfig) showPackage(btManager *bintray.ServicesManager, repo *Repo, pack *Package) error {
	pkg := fmt.Sprintf("%s/%s/%s", repo.Subject, repo.Name, pack.Package)
	fmt.Printf("\nPackage details.. [%s]", pkg)
	pkgPath, _ := packages.CreatePath(pkg)
	err := btManager.ShowPackage(pkgPath)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (config *BintrayConfig) checkDetails(btManager *bintray.ServicesManager) {
	var err error
	for _, repo := range config.Bintray.Repos {
		for _, pack := range repo.Packages {
			pkg := fmt.Sprintf("%s/%s/%s", repo.Subject, repo.Name, pack.Package)
			fmt.Printf("\nChecking details.. [%s]", pkg)

			// Repository
			RepoExistsOk, _ := btManager.IsRepoExists(&repositories.Path{Subject: repo.Subject, Repo: repo.Name})
			repoPath := fmt.Sprintf("%s/%s", repo.Subject, repo.Name)
			if RepoExistsOk != true {
				fmt.Printf("\nRepo does not exist.. [%s]", repoPath)
				fmt.Printf("\nCreating Repo..")
				err = config.createRepo(btManager, &repo)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else {
				fmt.Printf("\nRepo already exists.. [%s]", repoPath)
			}

			// Package
			pkgPath, _ := packages.CreatePath(pkg)
			PackageExistsOk, _ := btManager.IsPackageExists(pkgPath)
			if PackageExistsOk != true {
				fmt.Printf("\nPackage [%s] does not exist", pkg)
				fmt.Printf("\nCreating Package..\n")
				err = config.createPackage(btManager, &repo, &pack)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else {
				fmt.Printf("\nPackage already exists.. [%s]\n", pkg)
			}
		}
	}
}

func (config *BintrayConfig) cleanup(btManager *bintray.ServicesManager) {
	var err error
	for _, repo := range config.Bintray.Repos {
		for _, pack := range repo.Packages {
			err = config.deletePackage(btManager, &repo, &pack)
			if err != nil {
				fmt.Print(err)
			} else {
				config.deleteRepo(btManager, &repo)
			}
		}
	}
}
