# 03 · Khách Sạn & Phòng

[← Wiki](./README.md)

---

## Khách sạn (Hotel)

### Thông tin cơ bản

Mỗi khách sạn có:

- **Tên, mô tả, địa chỉ, thành phố** — dùng để tìm kiếm (full-text search)
- **Đánh giá (Rating):** thang 0–5 sao, hệ thống khởi tạo = 0 khi tạo mới
- **Giờ check-in / check-out:** nhập dạng `HH:MM` hoặc `HH:MM:SS`
- **Tuổi trẻ em tối đa (DefaultChildMaxAge):** mặc định **11 tuổi**. Trẻ từ 12 tuổi trở lên được tính là người lớn khi phân bổ phòng
- **Hình ảnh:** danh sách ảnh, có thể đánh dấu 1 ảnh là ảnh bìa
- **Phương thức thanh toán:** cấu hình cho từng khách sạn

### Phương thức thanh toán

| Giá trị        | Ý nghĩa                  |
| -------------- | ------------------------ |
| `immediate`    | Thanh toán ngay khi đặt  |
| `pay_at_hotel` | Thanh toán tại khách sạn |
| `deferred`     | Thanh toán trả sau       |

Mỗi phương thức có thể được bật/tắt (`enabled: true/false`). Khi tìm kiếm, user có thể lọc theo phương thức thanh toán mong muốn.

### Upload ảnh khách sạn

Ảnh được upload riêng lên S3, sau đó URL ảnh được dùng khi tạo khách sạn:

```
Bước 1: POST /api/hotels/upload-images  → nhận về URL
Bước 2: POST /api/hotels  → gắn URL ảnh vào payload
```

- Dung lượng tối đa: **10 MB/ảnh**
- Định dạng: JPEG, PNG, GIF, WebP, BMP
- Tên file được sinh tự động (timestamp + random)

---

## Phòng (Room)

### Thông tin cơ bản

Mỗi phòng là một **loại phòng** thuộc một khách sạn, có:

- **Tên, mô tả** — mô tả loại phòng
- **Giá cơ bản (BasePrice):** giá theo đêm, phải lớn hơn 0
- **Sức chứa:**
  - `MaxAdult`: số người lớn tối đa (phải ≥ 1)
  - `MaxChild`: số trẻ em tối đa (có thể = 0)
  - `MaxOccupancy`: tổng số người tối đa (phải ≥ MaxAdult + MaxChild)
- **Diện tích (SizeSqm):** m²
- **Tùy chọn giường (BedOptions):** lưu dạng JSON tự do
- **Tiện nghi (Amenities):** liên kết với danh mục tiện nghi
- **Hình ảnh:** URL ảnh nhúng trực tiếp khi tạo phòng (không có endpoint upload riêng)

> **Lưu ý:** Ảnh phòng khác ảnh khách sạn — không upload qua S3 endpoint. URL ảnh được cung cấp trực tiếp trong body khi gọi `POST /api/rooms`.

### Trạng thái phòng

| Trạng thái | Ý nghĩa                                         |
| ---------- | ----------------------------------------------- |
| `active`   | Phòng đang hoạt động, xuất hiện trong tìm kiếm  |
| `inactive` | Phòng tạm ngưng, không xuất hiện trong tìm kiếm |

Khi tạo mới, trạng thái mặc định là `active`.

---

## Tiện nghi phòng (Room Amenity)

Là danh mục chung, tạo một lần, dùng lại cho nhiều phòng:

- **Code:** mã định danh duy nhất (vd: `wifi`, `pool`, `breakfast`)
- **Tên:** tên hiển thị
- **Icon:** URL icon

Khi tạo phòng, cung cấp danh sách `amenityIds` để liên kết.  
Khi tìm kiếm, user có thể lọc phòng theo tiện nghi yêu cầu.

---

## Tồn kho phòng (Room Inventory)

Đây là phần quan trọng nhất của module phòng.

### Khái niệm

Mỗi bản ghi tồn kho xác định **số lượng phòng của loại đó trong một ngày cụ thể**:

| Trường            | Ý nghĩa                                          |
| ----------------- | ------------------------------------------------ |
| `TotalInventory`  | Tổng số phòng vật lý                             |
| `HeldInventory`   | Số phòng đang tạm giữ (chờ thanh toán)           |
| `BookedInventory` | Số phòng đã đặt xong                             |
| **Available**     | = Total − Held − Booked _(tính toán, không lưu)_ |

### Quy tắc tồn kho

- Inventory được tạo **thủ công cho từng ngày** qua API
- Một phòng **chỉ xuất hiện trong kết quả tìm kiếm** nếu có inventory với Available > 0 cho **tất cả các ngày** trong khoảng check-in đến check-out
- Nếu thiếu inventory cho bất kỳ ngày nào → phòng đó bị loại khỏi kết quả

**Ví dụ:** Tìm phòng từ 01/04 đến 03/04 (2 đêm: 01/04 và 02/04):

- Nếu 01/04 có Available = 2, 02/04 có Available = 0 → phòng **không** xuất hiện
- Nếu cả hai ngày đều Available > 0 → phòng **xuất hiện**, AvailableCount = min(2 ngày)

---

## Các API liên quan

| Method | Endpoint                     | Mô tả                            |
| ------ | ---------------------------- | -------------------------------- |
| GET    | `/api/hotels`                | Danh sách khách sạn              |
| GET    | `/api/hotels/:id`            | Chi tiết khách sạn               |
| POST   | `/api/hotels`                | Tạo khách sạn mới                |
| POST   | `/api/hotels/upload-images`  | Upload ảnh khách sạn lên S3      |
| POST   | `/api/rooms`                 | Tạo loại phòng                   |
| POST   | `/api/room-amenities`        | Tạo tiện nghi                    |
| POST   | `/api/rooms/:id/inventories` | Thêm tồn kho cho phòng theo ngày |
