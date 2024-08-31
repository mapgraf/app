Тестирование и аудит плагина карт Mapgl App (платная версия)

1.) https://github.com/mapgraf/builder войти под учетной записью:
username: <b>mapgl-user</b>
password: <b>4HH5W#*c</b>

2.) ознакомиться с исходным кодом. 

Серверный код в папке /pkg отвечает за синхронизацию изменений на карте между браузерной БД и файлом SQLite на сервере по адресу <b>/usr/share/grafana/public/seed/mapgl-data.db</b>. 
В ходе сборки папка /pkg компилируется в исполняемый файл <b>gpx_mapglapp_linux_amd64</b>, который Grafana запускает как часть плагина. 

Файлы клиентской части находятся в папке dist в минифицированном виде. Запускаются в браузере, и не имеют доступа к непубличным файлам на сервере. 

3.) запустить сборку:
 - перейти в раздел Actions;
 - выбрать процесс 'Build and release';
 - Run workflow;
 - указать название организации, выбрать ОС сервера, указать домен и порт, на котором работает Grafana (для выпуска сертификата плагина). 
 
   ![gh-actions](https://github.com/user-attachments/assets/b207f89b-af09-46fc-983b-5245b832eaba)

     
   В Grafana встроена проверка подлинности плагина. В связи с этим уточните настройки в разделе [server] вашего файла конфигурации grafana.ini https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana:
   
    "domain" - если конфигурировали этот пункт, значение должно совпадать с указанным выше доменом для сертификации плагина. Если нет - сертификат можно выпустить на адрес http://localhost:3000
   
    "allow_loading_unsigned_plugins" - если разрешен запуск неподписанных плагинов, сгенерированный файл сертификата MANIFEST.txt нужно удалить из папки плагина, иначе плагин не зарегистрируется.
    

4.) по окончании сборки (зеленая галочка напротив запущенного процесса 'Build and release') вернуться на главную страницу репозитория, и в разделе Releases скачать полученный архив с плагином

![releases](https://github.com/user-attachments/assets/9c1cbebd-38c8-47f1-9b84-821d4be34f0f)

5.) распаковать архив в папку с плагинами (по умолчанию /var/lib/grafana/plugins)

6.) Перезапустить сервер, в разделе App главного меню Grafana появится плагин Mapgl App. Перейдите в настройки и укажите лицензионный токен.

![map-settings](https://github.com/user-attachments/assets/03507bc4-77b1-429a-a8d9-61f2dd60487b)

Токен действует до: <b>2024-09-21 13:33:10 UTC</b>

```eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJvcmdOYW1lIjoiZGVtby1PcmciLCJkb21haW4iOiJkZW1vLm9yZyIsImZlYXRMaW1pdCI6MTAwLCJleHAiOjE3MjY5Mjc1MzJ9._yc9qUZOTY6Q3novj6EVKgiRJCYb1YmvkxrSgmXHPBg```
