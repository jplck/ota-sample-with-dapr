FROM node:8-alpine
WORKDIR /usr/src/app
COPY . .
RUN npm install
EXPOSE 8000
CMD [ "node", "app.js" ]