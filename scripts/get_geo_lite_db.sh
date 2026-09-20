#!/bin/bash
url=https://git.io/GeoLite2-City.mmdb # Getting GeoLite city db for higher resolution

curl -L -o ./assets/geolitedb/GeoLite2-City.mmdb $url