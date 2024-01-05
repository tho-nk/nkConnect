`docker build . -t iainttho/nk-connect:0.0.1`

`docker tag iainttho/nk-connect:0.0.1 iainttho/nk-connect:latest`

`docker push iainttho/nk-connect:latest`

`docker push iainttho/nk-connect:0.0.1`

`docker run -p 9096:9096 --name nk-connect iainttho/nk-connect:latest`
