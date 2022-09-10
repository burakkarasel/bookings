# Fort Smythe

![Project Image](./Fort-Smythe.gif)

---

### Table of Contents

- [Description](#description)
- [How To Use](#how-to-use)
- [Author Info](#author-info)

---

## Project Summary

You can visit the website, check for rooms and their availability. If there are availability for the room you want you can make reservation!
This Website also has an admin dashboard that enables owners to block dates for reasons such as modification and also owners can check the details of the scheduled reservations!

## Technologies

### Main Technologies

- [Go](https://go.dev/)
- [PostgreSQL](https://www.postgresql.org/)

### Libraries

- [alexedwards/scs](htts://www.github.com/alexedwards/scs/v2)
- [asaskevich/govalidator](htts://www.github.com/asaskevich/govalidator)
- [go-chi/chi](htts://www.github.com/go-chi/chi)
- [jackc/pgconn](htts://www.github.com/jackc/pgconn)
- [jackc/pgx](htts://www.github.com/jackc/pgx/v4)
- [justinas/nosurf](htts://www.github.com/justinas/nosurf)
- [xhit/go-simple-mail](htts://www.github.com/xhit/go-simple-mail/v2)
- [x/crypto](htts://www.github.com/x/crypto)

[Back To The Top](#Fort-Smythe)

---

## How To Use

### Tools

- [Go](https://go.dev/dl/)
- [dbeaver](https://dbeaver.io/download/)
- [Soda CLI](https://gobuffalo.io/documentation/database/soda/)

### Setup Database

- Create Database

```
CREATE DATABASE <your database name>
```

- Create your own database.yml file and run in terminal

```
soda migrate
```

- For dropping migration run

```
soda migrate down
```

### Run tests

- To run all tests

```
go test -v -cover ./...
```

### Start App

- Start the app

```
go build -o bookings cmd/web/*.go && ./bookings -dbname=<your db name> -dbuser=<your user name> -dbpw=<your password> -cache=true -production=false
```

## Author Info

- Twitter - [@dev_bck](https://twitter.com/dev_bck)

[Back To The Top](#Fort-Smythe)
