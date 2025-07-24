# build stage
FROM node:18 AS build
WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile
COPY src/ ./src
RUN yarn build

# final stage
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
# expose 80 by default 