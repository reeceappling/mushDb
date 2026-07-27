# syntax=docker.io/docker/dockerfile:1
FROM node:26.5-alpine AS base

WORKDIR /app

# Private Base API URL should use http and not https. Internal services in the cluster typically communicate over http, and the api service will handle TLS termination if needed. If we set this to https, the Next.js server will try to connect to the api service using https, which will fail since the api service is not configured for https.
# Must be available at runtime, which is why it is built into all images during this build process.
# TODO: consider moving to only the last one...
ENV NEXT_PRIVATE_BASE_API_URL="http://api:8080"

# Rebuild the source code only when needed
FROM base AS builder

# in base: WORKDIR /app

# Install dependencies based on the preferred package manager
COPY package.json yarn.lock* package-lock.json* pnpm-lock.yaml* .npmrc* ./
# Omit --production flag for TypeScript devDependencies # TODO: used to be just: npm ci, could be npm ci --omit=dev
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm i; \
  # Allow install without lockfile, so example works even without Node.js installed locally
  else echo "Warning: Lockfile not found. It is recommended to commit lockfiles to version control." && yarn install; \
  fi

COPY src ./src
COPY public ./public
COPY next.config.ts .
COPY postcss.config.mjs .
COPY tailwind.config.ts .
COPY tsconfig.json .

# Environment variables must be present at build time
# https://github.com/vercel/next.js/discussions/14030
ARG MAIN_API_EXTERNAL_HOST=mush.appli.ng
ARG MAIN_API_EXTERNAL_PROTOCOL=https
ARG MAIN_API_EXTERNAL_PORT=443
ARG GOOGLE_CLIENT_ID=badIdInDockerfile
ENV NEXT_PUBLIC_GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
ENV NEXT_PUBLIC_DOMAIN: ${MAIN_API_EXTERNAL_HOST}
ENV NEXT_PUBLIC_PORT: ${MAIN_API_EXTERNAL_PORT}
ENV NEXT_PUBLIC_BASE_API_URL=${MAIN_API_EXTERNAL_PROTOCOL}://${MAIN_API_EXTERNAL_HOST}

# Next.js collects completely anonymous telemetry data about general usage. Learn more here: https://nextjs.org/telemetry
# Uncomment the following line to disable telemetry at build time # TODO: CONSIDER
# ENV NEXT_TELEMETRY_DISABLED 1

# Build Next.js based on the preferred package manager
RUN \
  if [ -f yarn.lock ]; then yarn build; \
  elif [ -f package-lock.json ]; then INLINE_RUNTIME_CHUNK=false npm run build; \
  elif [ -f pnpm-lock.yaml ]; then pnpm build; \
  else npm run build; \
  fi



# Note: It is not necessary to add an intermediate step that does a full copy of `node_modules` here

# Create Production image, copy all the files, and run next
# Start with the base image which contains all private env vars and the base workdir
FROM base AS runner

RUN apk --no-cache add curl lsof

# Don't run production as root, # TODO: ensure ok!
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs
USER nextjs

COPY --from=builder /app/public ./public

# Automatically leverage output traces to reduce image size # TODO: this!
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# Uncomment the following line to disable telemetry at run time # TODO: this?
# ENV NEXT_TELEMETRY_DISABLED 1

CMD ["node", "server.js", "-p", "3000", "-H", "web"]