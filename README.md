# winKidsTimer

Если вам по каким-либо причинам не подходит функционал семейного контроля в Windows 10, 
то [данная, относительно простая программа,](https://github.com/n404an/winKidsTimer/releases/latest) может вам помочь.

Я искал какой-нибудь планировщик, который позволит задать режим работы ПК, но всё сводилось к тому, что этому уделяли мало времени в разработке.
Пришлось написать самому. :)

Базовое. То, что нужно было на данный момент, а именно:
1. Задать желаемое кол-во интервалов работы ПК.
2. При попытке войти в систему в "нерабочее" время - автоматическое выключение.
3. Ориентир на ребёнка, который ещё не научился обходить простые ограничения коварных родителей.
4. Работает на Windows 10. Более ранние(7/XP) по идее тоже подойдут.

### Как работает?
1. Установки не требует. UAC при запуске ругнётся, одобрите 1 раз и всё. В сеть не ходит. Логи не пишет.
2. exe-файл нужно добавить в автозагрузку. 
Самый простой способ - когда вы под целевой учёткой, 
в Проводнике в адрес вставьте `shell:startup` для перехода в папку Автозагрузки. 
Сюда скопируйте exe-файл.
2. При первом запуске будет создана папка `%HOMEPATH%\.winKidTimer` и внутри yml-файл с базовыми настройками.
3. Соблюдая формат, можете редактировать как вам угодно настройки. Применятся при следующем запуске программы.

### Как отключить?
1. При базовых настройках у вас есть 30-60 сек для того, что бы войти открыть консоль(`WIN+R  cmd`). И набрать там `shutdown -a`. Нажать Enter.
Эта комманда отменит ближайшее отключение ПК.
2. Далее нужно отменить фоновый процесс `winKidTimer.exe` через Диспетчер задач
3. И убрать файл из Автозагрузки

### Какие могут быть проблемы?
1. Сломаете формат yml-файла. Решение простое - удалить файл. При перезапуске базовый файл создастся.
2. Не знаю, удивите. :)

Если кому-нибудь интересно развитие этого проекта - готов ["за еду и битки"](https://www.blockchain.com/btc/address/1DLZJxtDsFQw6XN8RaBPEYothHqeeWkR1M) дальше развивать. 
Пока что меня всё устраивает как есть. :)

