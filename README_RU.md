<div align="center">
  
  # 🌟 SNET Professional

  <p align="center">
    <a href="README.md">English</a> | 
    <a href="README_RU.md">Русский</a>
  </p>

  <img src="assets/logo.png" alt="SNET Logo" width="200">

  <p align="center">
    <img src="https://img.shields.io/github/v/release/sky-night-net/SNET?color=blue&include_prereleases" alt="release">
    <img src="https://img.shields.io/github/actions/workflow/status/sky-night-net/SNET/release.yml?branch=main" alt="build">
    <img src="https://img.shields.io/github/go-mod/go-version/sky-night-net/SNET" alt="go version">
    <img src="https://img.shields.io/github/license/sky-night-net/SNET?color=green" alt="license">
  </p>

  **Высокопроизводительная нативная панель управления VPN-инфраструктурой для Xray-core и протоколов безопасности.**
  
</div>

---

> [!IMPORTANT]  
> Проект SNET разработан для профессионального управления VPN-сетями. Работает как нативный бинарный файл, обеспечивая максимальную скорость и стабильность без контейнеризации.

---

## 🔥 Ключевые возможности

- **🎨 Современный UX**: Премиальный Glassmorphism-интерфейс с мгновенной локализацией **RU/EN**.
- **📊 Живая телеметрия**: Реал-тайм мониторинг (CPU, RAM, Трафик) с обновлением каждые 2 сек.
- **🛡️ Встроенный Firewall**: Управление правилами `iptables` прямо из удобного веб-интерфейса.
- **🚀 Поддержка протоколов**: AmneziaWG (v1/v2), OpenVPN XOR, VLESS, VMess, Trojan, Shadowsocks.
- **🔒 Reality & TLS**: Тонкая настройка современных методов маскировки трафика.
- **📦 Автономность**: Установка одним файлом без внешних зависимостей.

---

## 🚀 Быстрый запуск

Для установки или обновления SNET выполните следующую команду:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/sky-night-net/SNET/main/install.sh)
```

Для ознакомления с подробной документацией посетите наши разделы на **English (EN)** или **Русском (RU)**.

---

## 🛠️ Управление через CLI (`snet`)

| Команда | Описание |
| :--- | :--- |
| `snet start` | Запустить службу панели |
| `snet stop` | Остановить службу панели |
| `snet restart` | Перезагрузить панель и ядро |
| `snet status` | Проверить статус работы |
| `snet logs` | Просмотр логов в реальном времени |

---

<div align="center">
  <sub>Разработано на Go и React. С верой в свободный интернет. 🌌</sub>
</div>
