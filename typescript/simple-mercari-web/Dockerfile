FROM node:16-alpine
WORKDIR /app

RUN addgroup -S mercari && adduser -S trainee -G mercari
USER trainee

CMD ["node", "-v"]
