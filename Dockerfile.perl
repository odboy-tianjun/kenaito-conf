# Build stage
FROM perl:5.40 AS build

WORKDIR /src

# Copy client library
COPY kenaito-conf-perl-client/ ./kenaito-conf-perl-client/

# Install client dependencies
WORKDIR /src/kenaito-conf-perl-client
RUN cpanm --notest HTTP::Tiny JSON::PP

# Copy demo project
WORKDIR /src
COPY kenaito-conf-perl-client-demo/ ./kenaito-conf-perl-client-demo/

# Runtime stage
FROM perl:5.40-slim

WORKDIR /app
COPY --from=build /src/kenaito-conf-perl-client/lib /app/kenaito-conf-perl-client/lib
COPY --from=build /src/kenaito-conf-perl-client-demo/app.pl .

ENV PERL5LIB=/app/kenaito-conf-perl-client/lib

ENTRYPOINT ["perl", "app.pl"]
