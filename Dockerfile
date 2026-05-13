# Build stage
FROM node:20-slim AS builder
WORKDIR /app
COPY src/package*.json ./
RUN if [ -f package.json ]; then npm install; fi
COPY src/ .

FROM node:20-slim
WORKDIR /app
COPY --from=builder /app .
EXPOSE 3000
CMD ["node", "app.js"]
