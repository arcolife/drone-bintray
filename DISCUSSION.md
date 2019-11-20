## Aim

- discussion @ https://gitter.im/drone/plugins?at=5d7fd98572fe125111b77576

### Plugin Design

> execute for each upload - or for multi-upload in a single step load the configuration from file.

> best practice - a plugin with super-complex configuration should define its own configuration file
> Need to check if globs could make sense as well

> so a plugin could support both this:

```
- name: upload
  settings:
    token:
     from_secret: token
    bucket: foo
    files:
    - bar.tar.gz
    - baz.tar.gz

OR

- name: upload
  settings:
    import: path/to/file.yaml
    token:
     from_secret: token
```

### Changes required

> old plugins required a json payload via stdin while newer plugins are simply listening to env variables. The nested array had been a poor design decision, that should be replaced by multiple calls of the bintray plugin instead. so keep the parameters as simple as possible, best would be plain types like int, float, string, boolean or slices of these simple types.

> Execute the plugin for every different upload instead of the nested attributes
