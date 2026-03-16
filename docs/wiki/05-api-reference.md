# 05 · API Quick Reference

[← Wiki](./README.md)

> Tài liệu đầy đủ hơn: Swagger UI tại `http://localhost:8088/swagger/index.html`

---

## Phân quyền

| Loại       | Mô tả                                                                  |
| ---------- | ---------------------------------------------------------------------- |
| **Public** | Không cần token                                                        |
| **JWT**    | Cần header `Authorization: Bearer <access_token>` và email đã xác thực |

> **Lưu ý kiến trúc hiện tại:** Hầu hết API đang ở dạng Public (chưa có middleware xác thực). Trong production, cần thêm bảo vệ cho các route tạo/sửa dữ liệu (khách sạn, phòng...).

---

## Rate Limiting

| Loại      | Giới hạn              | Áp dụng cho                                         |
| --------- | --------------------- | --------------------------------------------------- |
| Global    | 20 request/giây       | Tất cả endpoint                                     |
| Sensitive | 5 request/phút per IP | Login, verify email, forgot/reset password, refresh |

---

## Auth

| Method | Path                          | Auth   | Mô tả                         |
| ------ | ----------------------------- | ------ | ----------------------------- |
| POST   | `/api/auth/register`          | Public | Đăng ký tài khoản mới         |
| POST   | `/api/auth/login`             | Public | Đăng nhập (email + password)  |
| POST   | `/api/auth/refresh`           | Public | Làm mới access token          |
| POST   | `/api/auth/logout`            | JWT    | Đăng xuất                     |
| GET    | `/api/auth/me`                | JWT    | Thông tin user đang đăng nhập |
| POST   | `/api/auth/verify-email/send` | Public | Gửi lại email xác thực        |
| POST   | `/api/auth/verify-email`      | Public | Xác thực email bằng token     |
| POST   | `/api/auth/forgot-password`   | Public | Yêu cầu reset mật khẩu        |
| POST   | `/api/auth/reset-password`    | Public | Đặt mật khẩu mới              |
| GET    | `/api/auth/google/login`      | Public | Bắt đầu đăng nhập Google      |
| GET    | `/api/auth/google/callback`   | Public | Callback từ Google            |

---

## User

| Method | Path                         | Auth   | Mô tả                       |
| ------ | ---------------------------- | ------ | --------------------------- |
| GET    | `/api/users`                 | Public | Danh sách tất cả users      |
| GET    | `/api/users/:id`             | Public | Chi tiết user theo ID       |
| GET    | `/api/users/by-email?email=` | Public | Tìm user theo email         |
| POST   | `/api/users`                 | Public | Tạo user mới                |
| PATCH  | `/api/users/:id/profile`     | Public | Cập nhật tên, số điện thoại |
| PATCH  | `/api/users/:id/password`    | Public | Đổi mật khẩu                |
| PATCH  | `/api/users/:id/deactivate`  | Public | Vô hiệu hóa tài khoản       |

---

## Hotel

| Method | Path                        | Auth   | Mô tả               |
| ------ | --------------------------- | ------ | ------------------- |
| GET    | `/api/hotels`               | Public | Danh sách khách sạn |
| GET    | `/api/hotels/:hotel_id`     | Public | Chi tiết khách sạn  |
| POST   | `/api/hotels`               | Public | Tạo khách sạn mới   |
| POST   | `/api/hotels/upload-images` | Public | Upload ảnh lên S3   |

---

## Room

| Method | Path                              | Auth   | Mô tả                            |
| ------ | --------------------------------- | ------ | -------------------------------- |
| POST   | `/api/rooms`                      | Public | Tạo loại phòng                   |
| POST   | `/api/room-amenities`             | Public | Tạo tiện nghi                    |
| POST   | `/api/rooms/:room_id/inventories` | Public | Thêm tồn kho cho phòng theo ngày |

---

## Search

| Method | Path                                             | Auth   | Mô tả                         |
| ------ | ------------------------------------------------ | ------ | ----------------------------- |
| POST   | `/api/search/hotels`                             | Public | Tìm khách sạn (có phân trang) |
| POST   | `/api/search/hotels/:hotel_id/rooms`             | Public | Xem phòng available           |
| POST   | `/api/search/hotels/:hotel_id/room-combinations` | Public | Gợi ý tổ hợp phòng            |

---

## System

| Method | Path         | Mô tả        |
| ------ | ------------ | ------------ |
| GET    | `/health`    | Health check |
| GET    | `/swagger/*` | Swagger UI   |

---

## Response format

**Thành công:**

```json
{ "data": { ... } }
```

**Lỗi:**

```json
{
  "code": "invalid",
  "message": "user: email already exists",
  "info": "chi tiết lỗi (chỉ hiển thị ở môi trường development)"
}
```

**Mã lỗi:**

| Code              | HTTP Status | Khi nào                               |
| ----------------- | ----------- | ------------------------------------- |
| `invalid`         | 400         | Dữ liệu đầu vào sai                   |
| `unauthorized`    | 401         | Chưa xác thực hoặc token không hợp lệ |
| `not_found`       | 404         | Không tìm thấy resource               |
| `conflict`        | 409         | Trùng lặp (vd: email đã tồn tại)      |
| `internal`        | 500         | Lỗi server                            |
| `not_implemented` | 501         | Chức năng chưa làm                    |

---

## Request mẫu

### Đăng nhập

```json
POST /api/auth/login
{
  "email": "user@example.com",
  "password": "Password@123"
}
```

### Tạo khách sạn

```json
POST /api/hotels
{
  "name": "Khách Sạn Hà Nội",
  "address": "1 Hoàn Kiếm",
  "city": "Hà Nội",
  "checkInTime": "14:00",
  "checkOutTime": "12:00",
  "defaultChildMaxAge": 11,
  "images": [
    { "url": "https://cdn.example.com/hotel.jpg", "isCover": true }
  ],
  "paymentOptions": [
    { "paymentOption": "immediate", "enabled": true },
    { "paymentOption": "pay_at_hotel", "enabled": true }
  ]
}
```

### Tạo phòng

```json
POST /api/rooms
{
  "hotelId": "uuid-hotel",
  "name": "Deluxe Twin Room",
  "basePrice": 1200000,
  "maxAdult": 2,
  "maxChild": 1,
  "maxOccupancy": 3,
  "sizeSqm": 35,
  "images": [{ "url": "https://cdn.example.com/room.jpg", "isCover": true }],
  "amenityIds": ["uuid-amenity-1"]
}
```

### Thêm tồn kho phòng

```json
POST /api/rooms/:room_id/inventories
{
  "date": "2026-04-01",
  "totalInventory": 10,
  "heldInventory": 0,
  "bookedInventory": 0
}
```

### Tìm kiếm khách sạn

```json
POST /api/search/hotels
{
  "query": "hà nội",
  "checkInAt": "2026-04-01",
  "checkOutAt": "2026-04-03",
  "adultCount": 2,
  "childrenAges": [5],
  "roomCount": 1,
  "ratingMin": 3.0,
  "paymentOptions": ["immediate"],
  "page": 1,
  "pageSize": 10
}
```

### Tìm tổ hợp phòng

```json
POST /api/search/hotels/:hotel_id/room-combinations
{
  "checkInAt": "2026-04-01",
  "checkOutAt": "2026-04-03",
  "adultCount": 5,
  "childrenAges": [5, 8],
  "roomCount": 2,
  "maxCombinations": 5
}
```
