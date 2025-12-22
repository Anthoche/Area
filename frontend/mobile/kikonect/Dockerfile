FROM ghcr.io/cirruslabs/flutter:latest

WORKDIR /app

COPY pubspec.* ./

RUN flutter pub get

COPY . .

# Génération manuelle de local.properties pour s'assurer que les variables SDK sont définies
# Cela résout souvent les problèmes de plugins qui ne trouvent pas la configuration Flutter dans Docker
RUN echo "flutter.sdk=$(dirname $(dirname $(which flutter)))" > android/local.properties && \
    echo "flutter.buildMode=debug" >> android/local.properties && \
    echo "flutter.versionName=1.0.0" >> android/local.properties && \
    echo "flutter.versionCode=1" >> android/local.properties && \
    echo "flutter.compileSdkVersion=34" >> android/local.properties && \
    echo "flutter.minSdkVersion=19" >> android/local.properties && \
    echo "flutter.targetSdkVersion=34" >> android/local.properties

RUN yes | flutter doctor --android-licenses || true

# Nettoyage et build
RUN flutter clean
RUN flutter build apk --debug --no-pub || true

CMD ["echo", "Container mobile prêt. En attente de la commande de build..."]
