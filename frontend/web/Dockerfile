FROM node:latest as build

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

CMD ["npm", "run", "build"]

FROM nginx:alpine as run

WORKDIR /app

COPY --from=build ./dist /usr/share/nginx/html
