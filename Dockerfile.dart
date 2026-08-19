FROM dart:3.5

WORKDIR /kenaito-conf-dart-client-demo

# Copy client library
COPY kenaito-conf-dart-client/ /kenaito-conf-dart-client/

# Copy demo
COPY kenaito-conf-dart-client-demo/app.dart /kenaito-conf-dart-client-demo/app.dart

ENTRYPOINT ["dart", "run", "app.dart"]
