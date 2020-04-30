# Whatsapp web go rest 


- Para instalar
> Este projeto utiliza a linguagem de programação GO :
[Site da Linguagem](http://www.golangbr.org/)

## Estrutura do Projeto
```bash
├── api
│   ├── auth
│   │   └── token.go
│   ├── controllers
│   │   ├── base.go
│   │   ├── file_controller.go
│   │   ├── home_controller.go
│   │   ├── login_controller.go
│   │   ├── routes.go
│   │   ├── users_controller.go
│   │   └── whatsapp_controller.go
│   ├── helpers
│   │   └── config.go
│   ├── libs
│   │   ├── connhandler.go
│   │   └── whatsapp.go
│   ├── middlewares
│   │   └── middlewares.go
│   ├── models
│   │   ├── User.go
│   │   ├── UserToken.go
│   │   └── WpSession.go
│   ├── responses
│   │   └── response.go
│   ├── seed
│   │   └── seeder.go
│   ├── server.go
│   └── utils
│       └── formaterror
│           └── formaterror.go
├── go.mod
├── go.sum
├── main
├── main.go
└── share
    ├── etc
    │   ├── development.yaml
    │   └── production.yaml
    ├── private.key
    └── public.key
```

## Configuração do Projeto
> O projeto precisa de um banco de dados para ser executado 
>  - faça uma cópia do arquivo .env.example com o nome .env no diretório raiz do projeto executando o comando no terminal
>  ```bash
>   cp .env.example .env
>   ```

## Para Executar
> Após a configuração do ambiente go 
> No diretório raiz do projeto execute os comandos no terminal
> ```bash
> go build main.go
> go run main.go
>``` 


## Funcionamento da Aplicação

> Esta aplicação é uma api restfull, para auxilio no desenvolvimento é recomendado o uso de aplicações como Postman, ou Insomnia

#### Autenticação
> para poder realizar envio e recebimento de mensagens do whatsapp é necessário se autenticar primeiro na API.
>  - Token  de Autenticação
	>   a aplicação utiliza JWT como autentiação, para conseguir um token o usuário terá que fazer uma autnetiação básica via HTTP para o servidor
> ```bash 
>	curl -X POST \
>	  http://127.0.0.1:3000/login \
>	  -H 'Content-Type: application/json' \
>	  -d '{"email" : "admin@root.com", "password" : "root"}'
> ```  
>  A resposta para a requisição será o token semelhante 
> 
> ```bash
> "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJle HAiOjAsInVzZXJfaWQiOjJ9.sGDXEoKbCzWMHa9m-DPuC_BvUg8JgqqnQkVv2AQOzHI"
> ```

#### Autenticação com o whatsapp web
> Com o token em mãos, faça a autenticação para acesso ao whats app
> o parâmetro `output` é opicional para que a resposta da requisição sejá em HTML
> ```bash
> curl -X POST \
>  'http://127.0.0.1:3000/wp/login?output=html' \
>  -H 'Authorization: Bearer [TOKEN] \
>  -H 'Content-Type: application/json' \
> ```
> caso o parâmentro não esteja presente a resposta será semelhante a 
>```json 
>{
> "status": true,
> "code": 200,
>    "message": "Success",
>    "data": {
>        "qrcode": "data:image/png;base64,[código qr em base_64]",
>       "timeout": 5
>  }
>}
>```
>  - Com o parâmetro `output=html` a resposta será :
> ```html
> <html>
>    <head>
>       <title>WhatsApp Login</title>
>  </head>
>    <body>
>        <img src="data:image/png;base64,[código qr em base_64]" />
>        <p>
>            <b>QR Code Scan</b>
>            <br/>
>              Timeout in 5 Second(s)
>        </p>
>    </body>
></html>
>```
> O código tem validade de 5 segundo é deve ser lido pelo aplicativo como a opção de whatsapp web
 
