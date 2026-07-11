FROM node:20-alpine

WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
ARG PUBLIC_API_BASE=
ENV PUBLIC_API_BASE=${PUBLIC_API_BASE}
RUN npm run build
EXPOSE 5173
CMD ["npm", "run", "preview", "--", "--host", "0.0.0.0", "--port", "5173"]
