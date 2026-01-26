# Используем официальный образ NGINX
FROM nginx:latest

# Удаляем стандартный конфиг NGINX
RUN rm /etc/nginx/conf.d/default.conf

# Копируем наши конфигурационные файлы
COPY nginx.conf /etc/nginx/nginx.conf

# Открываем порт 80
EXPOSE 80

# Запускаем NGINX
CMD ["nginx", "-g", "daemon off;"]