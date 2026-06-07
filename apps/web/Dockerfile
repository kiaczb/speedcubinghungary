# 1. stage: Build
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install


ARG NEXT_PUBLIC_WCA_CLIENT_ID
ARG NEXT_PUBLIC_API_URL

# Átadjuk a build környezetnek (hogy a Next.js bele tudja égetni a kódba)
ENV NEXT_PUBLIC_WCA_CLIENT_ID=$NEXT_PUBLIC_WCA_CLIENT_ID
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

# EZZEL a sorral egyből látni fogod a terminálban buildeléskor, hogy megérkezett-e!
RUN echo "WCA Client ID is: $NEXT_PUBLIC_WCA_CLIENT_ID"

COPY . .
RUN npm run build

# 2. stage: Runtime
FROM node:22-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

# Csak a szükséges fájlokat másoljuk át
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json

EXPOSE 3000
CMD ["npm", "start"]