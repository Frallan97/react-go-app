# build stage
FROM oven/bun:1 AS build
WORKDIR /app
COPY package.json bun.lockb ./
RUN bun install --frozen-lockfile
COPY src/ ./src
COPY index.html ./
RUN bun run build

# final stage
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
# expose 80 by default 