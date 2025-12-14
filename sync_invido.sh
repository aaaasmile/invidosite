#!/bin/bash
# aggiornare post, pages e tutta la parte statica compreso js:
rsync -azv /mnt/d/Projects/go-lang/invido-site/invido-site/static/invido/ igor@invido.it:/var/www/invido.it/html

