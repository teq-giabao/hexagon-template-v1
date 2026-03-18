# 05 · API Quick Reference

[← Wiki](./README.md)

> Tài liệu đầy đủ hơn: Swagger UI tại `http://localhost:8088/swagger/index.html`

---

## Phân quyền

| Loại       | Mô tả                                                                  |
| ---------- | ---------------------------------------------------------------------- |
| **Public** | Không cần token                                                        |
| **JWT**    | Cần header `Authorization: Bearer <access_token>` và email đã xác thực |

> **Ghi chú:** Các route CRUD khách sạn và phòng hiện tại chưa yêu cầu JWT (dành cho admin tool). Trong production cần thêm role-based middleware cho admin routes.

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

| Method | Path                         | Auth | Mô tả                       |
| ------ | ---------------------------- | ---- | --------------------------- |
| GET    | `/api/users`                 | JWT  | Danh sách tất cả users      |
| GET    | `/api/users/:id`             | JWT  | Chi tiết user theo ID       |
| GET    | `/api/users/by-email?email=` | JWT  | Tìm user theo email         |
| POST   | `/api/users`                 | JWT  | Tạo user mới                |
| PATCH  | `/api/users/:id/profile`     | JWT  | Cập nhật tên, số điện thoại |
| PATCH  | `/api/users/:id/password`    | JWT  | Đổi mật khẩu                |
| PATCH  | `/api/users/:id/deactivate`  | JWT  | Vô hiệu hóa tài khoản       |

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

## Booking

| Method | Path                                | Auth | Mô tả                           |
| ------ | ----------------------------------- | ---- | ------------------------------- |
| POST   | `/api/bookings`                     | JWT  | Tạo booking, hold phòng 10 phút |
| GET    | `/api/bookings/:id`                 | JWT  | Xem chi tiết booking            |
| POST   | `/api/bookings/:id/payment-option`  | JWT  | Chọn phương thức thanh toán     |
| POST   | `/api/bookings/:id/confirm-payment` | JWT  | Xác nhận thanh toán thành công  |
| POST   | `/api/bookings/:id/cancel`          | JWT  | Hủy booking                     |

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

### Tạo booking

```json
POST /api/bookings
Authorization: Bearer <token>
{
  "roomId": "uuid-room",
  "checkInAt": "2026-04-01",
  "checkOutAt": "2026-04-03",
  "roomCount": 1,
  "guestCount": 2
}
```

### Chọn phương thức thanh toán

```json
POST /api/bookings/:id/payment-option
Authorization: Bearer <token>
{ "paymentOption": "immediate" }
```

### Xác nhận thanh toán thành công

```json
POST /api/bookings/:id/confirm-payment
Authorization: Bearer <token>
```

### Hủy booking

```json
POST /api/bookings/:id/cancel
Authorization: Bearer <token>
```
