# ---- builder ----
FROM node:20-alpine AS builder
ARG VITE_API_BASE=/api
ENV VITE_API_BASE=${VITE_API_BASE}
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci || npm install
COPY . .
RUN echo "VITE_API_BASE=${VITE_API_BASE}" > .env.production
RUN npm run build

# ---- runtime ----
FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
