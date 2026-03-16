# 01 · Tổng Quan Hệ Thống

[← Wiki](./README.md)

---

## Hệ thống này làm gì?

Backend API cho nền tảng đặt phòng khách sạn. Cho phép:

- Khách hàng tìm kiếm khách sạn, xem phòng có sẵn, chọn tổ hợp phòng phù hợp
- Admin quản lý khách sạn, phòng, tồn kho
- Người dùng đăng ký, đăng nhập, quản lý tài khoản

**Trạng thái hiện tại:** Hệ thống đang ở giai đoạn nền tảng. Chức năng đặt phòng thực tế (tạo booking, thanh toán) chưa được implement — luồng search đã hoàn chỉnh đến bước user chọn tổ hợp phòng.

---

## Tech stack

| Thành phần  | Công nghệ                    | Ghi chú                           |
| ----------- | ---------------------------- | --------------------------------- |
| Ngôn ngữ    | Go 1.24+                     |                                   |
| HTTP        | Echo framework               |                                   |
| Database    | PostgreSQL 15                |                                   |
| Auth        | JWT (access + refresh token) |                                   |
| OAuth       | Google OAuth 2.0             |                                   |
| Storage ảnh | AWS S3 / LocalStack          | LocalStack dùng cho dev local     |
| Email       | Resend API                   | Gửi mail xác thực, reset mật khẩu |
| Monitoring  | Sentry                       | Tự động báo lỗi 5xx               |

---

## Cấu trúc project

Project theo kiến trúc **Hexagonal** (còn gọi là Clean Architecture). Nguyên tắc chính: **business logic không phụ thuộc vào framework hay database**.

```
hexagon-template/
│
├── auth/          Nghiệp vụ xác thực (đăng ký, login, token...)
├── user/          Nghiệp vụ user (thông tin, mật khẩu...)
├── hotel/         Nghiệp vụ khách sạn
├── room/          Nghiệp vụ phòng + tồn kho
├── search/        Nghiệp vụ tìm kiếm + phân bổ phòng
├── upload/        Nghiệp vụ upload ảnh
│
├── httpserver/    HTTP handlers (nhận request, trả response)
├── postgres/      Truy vấn database
├── pkg/           Tiện ích dùng chung (JWT, bcrypt, S3, email...)
│
├── cmd/
│   ├── httpserver/  Entry point chạy server
│   ├── migrate/     Tool chạy migration database
│   └── seed/        Tool seed data mẫu
│
└── migrations/    File SQL tạo bảng
```

**Cách đọc code khi debug:** Bắt đầu từ `httpserver/` để tìm endpoint → vào domain tương ứng để xem business logic → vào `postgres/` để xem query.

---

## Chạy project local

```bash
make local-db    # Khởi động PostgreSQL (Docker, port 33062)
make db/migrate  # Tạo bảng
make run         # Chạy server với hot-reload (port 8088)
make test        # Chạy toàn bộ test
make lint        # Kiểm tra code style
```

Swagger UI: `http://localhost:8088/swagger/index.html`
