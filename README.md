# jwt

# Запуск 

1. Запустить Docker Desktop
2. Прописать в CLI: docker compose up -d
3. Если интересно, то можно смотреть логи: docker logs medods-auth-app-1 (либо с параметром -f)

**Используемые технологии:**

- Go
- JWT
- MongoDB

**Задание:**

Написать часть сервиса аутентификации.

Два REST маршрута:

- Первый маршрут выдает пару Access, Refresh токенов для пользователя с идентификатором (GUID) указанным в параметре запроса
- Второй маршрут выполняет Refresh операцию на пару Access, Refresh токенов

**Требования:**

Access токен тип JWT, алгоритм SHA512, хранить в базе строго запрещено.

Refresh токен тип произвольный, формат передачи base64, хранится в базе исключительно в виде bcrypt хеша, должен быть защищен от изменения настороне клиента и попыток повторного использования.

Access, Refresh токены обоюдно связаны, Refresh операцию для Access токена можно выполнить только тем Refresh токеном который был выдан вместе с ним.

**Результат:**

Результат выполнения задания нужно предоставить в виде исходного кода на Github.

## Как протестировать

1. Перейти в Postman
2. Прописать: 
   * Либо GET-запрос на localhost:8080/auth с указанием хедера Name (выбранный вариант параметра для аунтификации). Ответ: 
   1. Параметр аунтификации
   2. JWT-token лишь для наглядности
   3. Незашифрованный refresh-token
   
   ### Что происходит: 
   1. В куки устанавливается jwt и refresh-token устанавливаются в качестве инструкции к куки на стороне клиента по пути /api/auth. JWT в body, Refresh строго в HttpOnly, также устанавливаются их ttl
   2. В базу заноситься параметр аутификации, зашифрованный в bcrypt refresh-token и момент создания refresh-token
   3. Стоит заметить, что процесс добавления сессии в таблицу имеет свои меры безопасности. При добавлении проверятся сколько рефреш-сессий всего есть у юзера и, если их больше одной (для примера) или юзер конектится одновременно из нескольких подсетей, стоит предпринять меры. Имплементируя данную проверку, я проверяю только что бы юзер имел максимум 1 одновременных рефреш-сессий, и при попытке установить следующую удаляю предыдущую.
   Таким образом если юзер залогинился на пяти устройствах, рефреш токены будут постоянно обновляться и все счастливы. Но если с аккаунтом юзера начнут производить подозрительные действия(попытаются залогинится более чем на одном устройстве) система сбросит все сессии(рефреш токены).

   * Либо POST-запрос на localhost:8080/refresh с хедерами Name (тот же, на который получен токен) и Token (полученный ранее токен). Ответ:
   1. Параметр аунтификации
   2. JWT-token лишь для наглядности
   3. Незашифрованный refresh-token

   ### Что происходит: 
   1. Сверяются полученный токен и расшифрованный токен из базы
   2. Проверяю "жив" ли ещё токен
   3. В случае успеха обновляю рефреш-сессию в базе создав новый refresh-token и зашифровав его, также обновляю куки 

## Тонкости: 

1. Старался не использовать сторонних библиотек, для удволетворения "Используемые технологии", поэтому пришлось писать на net/http, вместо привычных фреймворков.
2. Можно протестировать без Докера, тогда нужно обновить config.env "CONFIG_PATH=./config/prod.yaml" -> "CONFIG_PATH=./config/local.yaml" и соответственно запустить MongoDB на локальной машине.
3. Нет управления "руками" JWT токенов (бан, удаление руками, т.к. нарушалась бы суть jwt токенов). Согласно: https://gist.github.com/zmts/802dc9c3510d79fd40f9dc38a12bccfc
4. config.env оставил лишь для наглядности

# Спасибо за внимание, буду рад фид-беку. 
*https://github.com/ZiganshinDev* 

# Демонстрация: 

1. Позитивный сценарий:
   ![Imgur](https://i.imgur.com/dmaJxjE.png)
   ![Imgur](https://i.imgur.com/TVvYYDN.png)
   ![Imgur](https://i.imgur.com/LU8AzJK.png)
   ![Imgur](https://i.imgur.com/CkgsIVf.png)

2. Негативный сценарий. Разные варианты:
   ![Imgur](https://i.imgur.com/xZSgTL9.png)
   ![Imgur](https://i.imgur.com/gUf8PZ8.png)
   ![Imgur](https://i.imgur.com/sWuZpfd.png)
   ![Imgur](https://i.imgur.com/PMRTdGB.png)

