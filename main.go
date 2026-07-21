package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Menu struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Type  string  `json:"type"`
}

var menus = []Menu{
	{ID: 1, Name: "ต้มยำกุ้ง", Price: 120, Type: "soup"},
	{ID: 2, Name: "พิซซ่าฮาวายเอี้ยน", Price: 199, Type: "pizza"},
	{ID: 3, Name: "พิซซ่าเห็ด", Price: 179, Type: "pizza"},
}

var nextID = 4

// ด่านที่ 5: Helper Function สำหรับจัดการ Error Format แบบมืออาชีพ
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// ด่านที่ 2 & 7: GET /menu (พร้อมระบบ Filter ด้วย Query Parameter ?type=...)
func handleMenus(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		menuType := r.URL.Query().Get("type")
		if menuType != "" {
			var filtered []Menu
			for _, m := range menus {
				if m.Type == menuType {
					filtered = append(filtered, m)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(filtered)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(menus)
		return
	}

	// ด่านที่ 4: POST /menu (รับสั่งจานใหม่)
	if r.Method == http.MethodPost {
		var m Menu
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
			return
		}
		if m.Name == "" || m.Price <= 0 {
			writeError(w, http.StatusBadRequest, "MISSING_FIELD", "ต้องมีชื่อเมนู และราคาต้องมากกว่าศูนย์")
			return
		}
		m.ID = nextID
		nextID++
		menus = append(menus, m)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(m)
		return
	}
}

// ด่านที่ 3 & 6: GET, PUT, DELETE /menu/{id}
func handleMenuByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขจานต้องเป็นตัวเลข")
		return
	}

	// GET /menu/{id} (ขอดูทีละจาน)
	if r.Method == http.MethodGet {
		for _, m := range menus {
			if m.ID == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(m)
				return
			}
		}
		writeError(w, http.StatusNotFound, "MENU_NOT_FOUND", "ไม่พบเมนูหมายเลขนี้")
		return
	}

	// PUT /menu/{id} (แก้ไขจาน)
	if r.Method == http.MethodPut {
		var updatedMenu Menu
		if err := json.NewDecoder(r.Body).Decode(&updatedMenu); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
			return
		}

		for i, m := range menus {
			if m.ID == id {
				updatedMenu.ID = id
				menus[i] = updatedMenu
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(updatedMenu)
				return
			}
		}
		writeError(w, http.StatusNotFound, "MENU_NOT_FOUND", "ไม่พบเมนูหมายเลขนี้")
		return
	}

	// DELETE /menu/{id} (ลบจาน)
	if r.Method == http.MethodDelete {
		for i, m := range menus {
			if m.ID == id {
				menus = append(menus[:i], menus[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		writeError(w, http.StatusNotFound, "MENU_NOT_FOUND", "ไม่พบเมนูหมายเลขนี้")
		return
	}
}

// ด่านที่ 8: Middleware สำหรับรองรับ CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /menu", handleMenus)
	mux.HandleFunc("POST /menu", handleMenus)
	mux.HandleFunc("GET /menu/{id}", handleMenuByID)
	mux.HandleFunc("PUT /menu/{id}", handleMenuByID)
	mux.HandleFunc("DELETE /menu/{id}", handleMenuByID)

	println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", corsMiddleware(mux))
}