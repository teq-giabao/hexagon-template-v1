# 04 · Tìm Kiếm & Đặt Phòng

[← Wiki](./README.md)

---

## Tổng quan luồng tìm kiếm

Tìm kiếm diễn ra theo 3 bước, mỗi bước đào sâu hơn:

```
Bước 1  →  Tìm khách sạn phù hợp
Bước 2  →  Xem danh sách phòng available của khách sạn đó
Bước 3  →  Gợi ý tổ hợp phòng tối ưu để đặt
```

---

## Bước 1 · Tìm khách sạn

**Endpoint:** `POST /api/search/hotels`

### Tiêu chí tìm kiếm

| Tiêu chí                       | Bắt buộc | Mô tả                                      |
| ------------------------------ | -------- | ------------------------------------------ |
| `query`                        | ✅       | Từ khóa — tìm theo tên, thành phố, địa chỉ |
| `checkInAt`                    | ✅       | Ngày check-in (YYYY-MM-DD)                 |
| `checkOutAt`                   | ✅       | Ngày check-out (YYYY-MM-DD)                |
| `adultCount`                   | ✅       | Số người lớn (≥ 1)                         |
| `roomCount`                    | ✅       | Số phòng cần đặt (≥ 1)                     |
| `childrenAges`                 |          | Danh sách tuổi trẻ em, mỗi tuổi từ 0–17    |
| `ratingMin`                    |          | Đánh giá tối thiểu (0–5)                   |
| `amenityIds`                   |          | Lọc theo tiện nghi                         |
| `paymentOptions`               |          | Lọc theo phương thức thanh toán            |
| `page` / `pageSize` / `offset` |          | Phân trang (mặc định 10/trang)             |

### Quy tắc ngày

- Không được là ngày trong quá khứ
- CheckOut phải sau CheckIn
- Phải nằm trong cửa sổ đặt phòng hợp lệ

### Kết quả trả về

Mỗi khách sạn trong kết quả có 2 cờ quan trọng:

| Cờ                 | Ý nghĩa                                                 |
| ------------------ | ------------------------------------------------------- |
| `matchesRequested` | Có đủ phòng và đáp ứng được toàn bộ yêu cầu về số người |
| `flexibleMatch`    | Có phòng available nhưng không đủ để đáp ứng đầy đủ     |

Sắp xếp: `matchesRequested` hiện trước, rồi theo giá từ thấp đến cao.

---

## Bước 2 · Xem phòng của một khách sạn

**Endpoint:** `POST /api/search/hotels/:hotel_id/rooms`

Trả về danh sách phòng available của khách sạn trong khoảng ngày đó, kèm:

- Số lượng phòng còn trống (`availableCount`)
- Danh sách tiện nghi
- `strictMatch`: tổng phòng available ≥ số phòng yêu cầu

---

## Bước 3 · Gợi ý tổ hợp phòng

**Endpoint:** `POST /api/search/hotels/:hotel_id/room-combinations`

Đây là tính năng quan trọng nhất: **tìm cách ghép phòng tối ưu** để đủ chỗ cho toàn bộ nhóm khách, với chi phí thấp nhất.

**Ví dụ:** Nhóm 5 người lớn + 1 trẻ 5 tuổi, cần 2 phòng:

```
Phòng A: tối đa 2 người lớn + 1 trẻ, giá 1.000.000đ
Phòng B: tối đa 3 người lớn + 2 trẻ, giá 1.500.000đ

Tổ hợp gợi ý:
  1. 2×A = 4 người lớn + 2 trẻ, tổng 2.000.000đ  ✅
  2. 1×A + 1×B = 5 người lớn + 3 trẻ, tổng 2.500.000đ  ✅
  3. 2×B = 6 người lớn + 4 trẻ, tổng 3.000.000đ  ✅
```

Tham số `maxCombinations` (mặc định 5): giới hạn số tổ hợp trả về.

---

## Quy tắc phân bổ tuổi trẻ em

> ⚠️ **Đây là quy tắc nghiệp vụ quan trọng — cần nắm rõ trước khi sửa code liên quan.**

### Xác định trẻ em hay người lớn

Mỗi khách sạn có `DefaultChildMaxAge` (mặc định: **11**).

- Trẻ có tuổi **≤ DefaultChildMaxAge** → tính là **trẻ em**, chiếm slot MaxChild
- Trẻ có tuổi **> DefaultChildMaxAge** → tính là **người lớn**, chiếm slot MaxAdult

**Ví dụ** với `DefaultChildMaxAge = 11`:

- Trẻ 10 tuổi → trẻ em
- Trẻ 12 tuổi → người lớn (tốn slot người lớn)

### Quy tắc: Trẻ em không được ở một mình

**Mỗi phòng có trẻ em phải có ít nhất 1 người lớn đi kèm.**

Quy tắc này được kiểm tra **ở cả hai nơi:**

1. Khi tính `matchesRequested` trong kết quả tìm kiếm khách sạn
2. Khi generate tổ hợp phòng

Ví dụ vi phạm:

```
Yêu cầu: 0 người lớn, 2 trẻ em, 1 phòng
→ Bị từ chối — không thể để trẻ ở một mình
```

Ví dụ hợp lệ:

```
Yêu cầu: 1 người lớn, 2 trẻ em, 2 phòng
→ Phân bổ: Phòng 1: 1 người lớn + 1 trẻ ✅ | Phòng 2: 0 người lớn + 1 trẻ ❌
→ Hệ thống thử phân bổ khác hoặc báo không đủ phòng
```

---

## Lưu ý khi phát triển tính năng đặt phòng

Tồn kho được cập nhật tự động trong mỗi bước booking:

1. `CreateBooking` → tăng `HeldInventory` cho các ngày trong khoảng đặt phòng
2. Chọn `pay_at_hotel` / `deferred` → chuyển `HeldInventory` → `BookedInventory` (confirm ngay)
3. Chọn `immediate` + thanh toán thành công → chuyển `HeldInventory` → `BookedInventory`
4. Hủy booking pending → giảm `HeldInventory`; hủy confirmed → giảm `BookedInventory`
5. Expire booking → tương tự hủy nhưng tự động theo thời gian

---

## Luồng đặt phòng (Booking Flow)

Sau khi tìm kiếm và chọn được tổ hợp phòng, user tiến hành đặt phòng theo các bước sau:

```
Bước 1  →  Tạo booking (POST /api/bookings)
Bước 2  →  Chọn phương thức thanh toán
Bước 3  →  Thanh toán (nếu cần)
Bước 4  →  Hủy booking (tùy chọn)
```

---

### Bước 1 · Tạo Booking

**Endpoint:** `POST /api/bookings` _(yêu cầu JWT)_

**Request:**

```json
{
  "roomId": "uuid-room",
  "checkInAt": "2026-04-01",
  "checkOutAt": "2026-04-03",
  "roomCount": 1,
  "guestCount": 2
}
```

**Hệ thống thực hiện (trong cùng một transaction):**

1. Lock room inventories bằng `SELECT FOR UPDATE` cho tất cả các ngày trong khoảng đặt
2. Kiểm tra phòng còn trống: `Available = Total − Held − Booked ≥ roomCount`
3. Nếu không đủ phòng → trả về lỗi `409 Conflict` — "room is not available"
4. Tăng `HeldInventory += roomCount` cho mỗi ngày
5. Tính `total_price = nightly_price × nights × roomCount`
6. Tạo bản ghi booking với trạng thái `pending`
7. Set `hold_expires_at = now + 10 phút`

**Concurrency handling:** Nhiều user cùng đặt phòng cuối cùng — chỉ request đầu tiên lock được inventory sẽ thành công. Các request còn lại nhận `409`.

**Response** trả về booking mới tạo + danh sách phương thức thanh toán mà khách sạn hỗ trợ:

```json
{
  "booking": {
    "id": "uuid-booking",
    "userId": "uuid-user",
    "hotelId": "uuid-hotel",
    "roomId": "uuid-room",
    "checkInAt": "2026-04-01T00:00:00Z",
    "checkOutAt": "2026-04-03T00:00:00Z",
    "nights": 2,
    "roomCount": 1,
    "guestCount": 2,
    "nightlyPrice": 1200000,
    "totalPrice": 2400000,
    "status": "pending",
    "paymentStatus": "unpaid",
    "holdExpiresAt": "2026-03-18T10:10:00Z"
  },
  "paymentOptions": ["immediate", "pay_at_hotel", "deferred"]
}
```

---

### Bước 2 · Chọn phương thức thanh toán

**Endpoint:** `POST /api/bookings/:id/payment-option` _(yêu cầu JWT)_

Chỉ được gọi khi booking đang ở trạng thái `pending` và chưa hết hold time.

```json
{ "paymentOption": "immediate" }
```

Hệ thống kiểm tra phương thức chọn phải nằm trong danh sách **enabled** của khách sạn đó. Nếu không → `400 Bad Request`.

#### Hành vi theo từng phương thức

| Phương thức    | Hành vi sau khi chọn                                            | Trạng thái mới |
| -------------- | --------------------------------------------------------------- | -------------- |
| `immediate`    | Giữ nguyên `pending`. Set `payment_deadline = hold_expires_at`  | `pending`      |
| `pay_at_hotel` | Confirm ngay, Held → Booked. Xóa `hold_expires_at`              | `confirmed`    |
| `deferred`     | Confirm ngay, Held → Booked. Set `payment_deadline = now + 24h` | `confirmed`    |

---

### Bước 3 · Thanh toán (chỉ áp dụng cho `immediate`)

**Endpoint:** `POST /api/bookings/:id/confirm-payment` _(yêu cầu JWT)_

Gọi sau khi payment gateway xác nhận thành công.

- Nếu booking đang `pending` và chưa hết hold → Held → Booked, chuyển sang `confirmed`
- Nếu booking đang `confirmed` (deferred) → cập nhật `payment_status = paid`

---

### Bước 4 · Hủy booking

**Endpoint:** `POST /api/bookings/:id/cancel` _(yêu cầu JWT)_

| Trạng thái hiện tại  | Hành vi                                                              |
| -------------------- | -------------------------------------------------------------------- |
| `pending`            | Giảm `HeldInventory`. Không hoàn tiền (chưa thanh toán)              |
| `confirmed` + paid   | Giảm `BookedInventory`. Tính hoàn tiền: `refund = total_price − fee` |
| `confirmed` + unpaid | Giảm `BookedInventory`. Không hoàn tiền                              |
| `cancelled`          | Lỗi: đã hủy                                                          |
| `expired`            | Lỗi: đã hết hạn                                                      |

> **Lưu ý về cancellation fee:** Mặc định là 0. Trong tương lai có thể implement policy theo số ngày trước check-in.

---

## Trạng thái booking (Booking Status)

```
         Tạo booking
              │
              ▼
          [pending]
         /    |    \
        /     |     \
   hết hold  chọn   cancel
   10 phút  payment  thủ công
        \     |     /
         \    ▼    /
          [expired] [cancelled]
              │
         pay_at_hotel
         hoặc deferred
              │
              ▼
          [confirmed]
          /        \
    thanh toán    hết payment_deadline
    thành công    (deferred unpaid)
         │              │
         ▼              ▼
    payment=paid    [expired]
```

### Bảng trạng thái đầy đủ

| Status      | Ý nghĩa                                        | Có thể hủy? |
| ----------- | ---------------------------------------------- | ----------- |
| `pending`   | Đang chờ chọn phương thức thanh toán; hold tạm | ✅ Có       |
| `confirmed` | Đã xác nhận; phòng được giữ chắc               | ✅ Có       |
| `cancelled` | Đã hủy                                         | ❌ Không    |
| `expired`   | Hết hạn tự động                                | ❌ Không    |

---

## Tự động expire booking

Booking tự động chuyển sang `expired` trong 2 trường hợp:

| Trường hợp                      | Điều kiện                                                       |
| ------------------------------- | --------------------------------------------------------------- |
| Quá hold time                   | `status = pending` AND `hold_expires_at <= now`                 |
| Không thanh toán kịp (deferred) | `status = confirmed` AND `payment_deadline <= now` AND `unpaid` |

**Khi expire:**

- `pending` → giảm `HeldInventory`
- `confirmed` → giảm `BookedInventory`

Expire được trigger tự động khi có bất kỳ request nào đến `POST /api/bookings` hoặc `GET /api/bookings/:id`.

---

## Các API liên quan

| Method | Endpoint                                   | Auth   | Mô tả                           |
| ------ | ------------------------------------------ | ------ | ------------------------------- |
| POST   | `/api/search/hotels`                       | Public | Tìm khách sạn (có phân trang)   |
| POST   | `/api/search/hotels/:id/rooms`             | Public | Xem phòng available             |
| POST   | `/api/search/hotels/:id/room-combinations` | Public | Gợi ý tổ hợp phòng              |
| POST   | `/api/bookings`                            | JWT    | Tạo booking, hold phòng 10 phút |
| GET    | `/api/bookings/:id`                        | JWT    | Xem chi tiết booking            |
| POST   | `/api/bookings/:id/payment-option`         | JWT    | Chọn phương thức thanh toán     |
| POST   | `/api/bookings/:id/confirm-payment`        | JWT    | Xác nhận thanh toán thành công  |
| POST   | `/api/bookings/:id/cancel`                 | JWT    | Hủy booking                     |
