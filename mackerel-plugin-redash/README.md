mackerel-plugin-redash
=====================

[re:dash](http://redash.io/) custom metrics plugin for [mackerel-agent](https://github.com/mackerelio/mackerel-agent).

## Synopsis

```shell
mackerel-plugin-redash -api-key=YOUR_API_KEY [-url=http://localhost:5000] [-metric-key-prefix=redash] [-timeout=5]
```

## Example of mackerel-agent.conf

```
[plugin.metrics.redash]
command = "/path/to/mackerel-plugin-redash -api-key=YOUR_API_KEY"
```

## Author
[ariarijp](https://github.com/ariarijp)
