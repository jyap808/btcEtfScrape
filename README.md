## Bitcoin ETF scraper

This service scrapes the Bitcoin ETF web sites and reports the information on the X account [@ubiqetfbot](https://twitter.com/ubiqetfbot).

## Set up

This is largely a personal project so I just use a wrapper script. Set up the variables accordingly.

runme.sh
```
#!/bin/bash

## X
export GOTWI_API_KEY=
export GOTWI_API_KEY_SECRET=
export GOTWI_ACCESS_TOKEN=
export GOTWI_ACCESS_TOKEN_SECRET=

./btcEtfScrape -webhookURL https://discord.com/api/webhooks/[SET THIS]
```

## TODO

Further Dockerize the set up.

## License

[MIT License](LICENSE)
