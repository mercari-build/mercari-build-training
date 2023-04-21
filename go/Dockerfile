FROM alpine

RUN addgroup -S mercari && adduser -S trainee -G mercari
USER trainee

CMD ["go", "version"]
