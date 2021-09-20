FROM golang:1.17 AS builder
WORKDIR /app
COPY . .
RUN make compile

FROM alpine:3.13
RUN mkdir view
COPY --from=builder /app/invoice .
COPY --from=builder /app/view/invoice_form.html view/
COPY --from=builder /app/view/invoice.html view/
RUN apk add --update --no-cache \
    libgcc libstdc++ libx11 glib libxrender libxext libintl \
    ttf-dejavu ttf-droid ttf-freefont ttf-liberation ttf-ubuntu-font-family
COPY --from=madnight/alpine-wkhtmltopdf-builder:0.12.6-alpine3.10 \
    /bin/wkhtmltopdf /bin/wkhtmltopdf
EXPOSE 8000
CMD ["./invoice"]
