FROM php:8.1-apache
RUN docker-php-ext-install mysqli
COPY src/ /var/www/html/
RUN chmod -R 755 /var/www/html