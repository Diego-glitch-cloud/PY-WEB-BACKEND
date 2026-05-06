FROM golang:1.26-alpine

WORKDIR /app

# Instalar dependencias necesarias para compilar y ejecutar
RUN apk add --no-cache gcc musl-dev

# Copiar archivos de modulo y descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Compilar la aplicación
RUN go build -o main .

# Exponer el puerto
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"]
