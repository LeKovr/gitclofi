# gitclofi
git clone &amp; filter

Подготовка файлов из git-репозиториев для генераторов статических сайтов.

## Входные данные

* есть несколько гит-реп, в которых есть файлы в md

## Задача

* по заданному списку репозиториев получить их актуальные версии
* для каждого репозитория по каждому файлу из списка
  * применить фильтр
  * добавить результат работы шаблона для заголовка/подвала
  * сохранить в заданное место

