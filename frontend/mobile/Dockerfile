FROM ghcr.io/cirruslabs/flutter:3.10.6 as build

WORKDIR /app

COPY pubspec.yaml pubspec.lock* ./
RUN flutter pub get

COPY . .

RUN flutter pub get

RUN flutter build apk --release

FROM scratch AS artifact
COPY --from=build /app/build/app/outputs/flutter-apk/app-release.apk /
