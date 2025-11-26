FROM ghcr.io/cirruslabs/flutter:3.24.3

WORKDIR /app

COPY pubspec.* ./

RUN flutter pub get

COPY . .

RUN yes | flutter doctor --android-licenses || true

RUN flutter build apk --debug --no-pub || true

CMD ["echo", "Container mobile prêt. En attente de la commande de build..."]
