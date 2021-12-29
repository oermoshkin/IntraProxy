# Информация

Данный сервис позволяет использовать собственный домен, без переадресаций на сторонние домены.

В config.yml необходимо изменить домены в разделе **proxy** на ваши.

**origin.server** - установите выданный вам домен вендором.

Для запуска сервиса необходимо иметь сертификат в котором указаны все домены перечисленные в конфигурационном файле 
в разделе **proxy** или иметь сертификат типа wildcard.

## Let's Encrypt

Если у вас нет сертификата, вы можете выпустить бесплатный сертификат на 90 дней от Let's Encrypt

Для этого Вам понадобиться приложение certbot

Debian/Ubuntu
```apt install certbot```

Centos/RH
```yum install certbot```

* Убедитесь что на 80 и 443 порту у вас ничего не запущено.
* FQDN указанные в config.yml ссылаются на ваш сервер.

Для выпуска сертификата запустите команду ниже, предварительно изменим доменные имена на свои:
```
certbot certonly --standalone -d customer.domain -d login.customer.domain -d apigw.customer.domain -d documentserver.customer.domain
```

В случае успеха, измените в config.yml пусть до созданного сертификата

```
  cert:    "/etc/letsencrypt/live/customer.domain/fullchain.pem"
  privkey: "/etc/letsencrypt/live/customer.domain/privkey.pem"
```

# Запуск
### Бинарный файл
**Нужен установленный golang - https://go.dev**

Из основной директории запустите: ```go run ./cmd```
Для компиляции бинарного файла: ```make build```
В папке bin появится файл **IntraProxy** для ОС Linux.
Для компиляции под другую архитектуру используйте команду: 

```GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o bin/IntraProxy ./cmd/```

Изменив ОС и Архитектуру

### Docker
Для запуска системы из контейнера, вы можете использовать команду:

```make docker```

Система запустит докер с go, в котором собирется бинарный файл и подготовит докер для запуска.

Сборку можно пропустить и использовать готовый контейнер, скачав его с https://hub.docker.com/:

```
docker pull oermoshkin/intraproxy
```

Для запуска контейнера выполните:

```
docker run -p 443:443 -v $(pwd)/ssl:/srv/IntraProxy/ssl -v $(pwd)/config.yml:/srv/IntraProxy/config.yml --rm --name IntraProxy oermoshkin/intraproxy:latest 
```

*Не забываем про сертификаты, в данном примере они должны быть в папке **ssl**

### Docker Compose

Для использования Docker Compose, он должен быть установлен на сервере.

```
docker-compose up
```

Для запуска в фоне, добавить флаг **-d**
Для просмотра логов в этом случае используем ```docker-compose logs -f```

