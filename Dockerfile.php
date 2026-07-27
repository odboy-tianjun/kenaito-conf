FROM php:8.3-cli

WORKDIR /kenaito-conf-php-client-demo

# Install required extensions
RUN docker-php-ext-install pcntl

# Copy client library
COPY kenaito-conf-php-client/ /kenaito-conf-php-client/

# Copy demo
COPY kenaito-conf-php-client-demo/app.php /kenaito-conf-php-client-demo/app.php

ENTRYPOINT ["php", "app.php"]
