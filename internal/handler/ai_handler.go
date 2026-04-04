package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// AskAI meneruskan pertanyaan ke Gemini AI dan mengembalikan jawabannya.
func AskAI(c *gin.Context) {
	var input struct {
		Question string `json:"question"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		c.JSON(500, gin.H{"error": "API key tidak dikonfigurasi"})
		return
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + apiKey
	prompt := fmt.Sprintf("Jawab sebagai asisten Roti 515: %s", input.Question)

	reqBody := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]string{"text": prompt},
				},
			},
		},
	}

	jb, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses permintaan"})
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jb))
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal menghubungi AI"})
		return
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	answer := "Maaf, AI sedang sibuk."
	if candidates, ok := res["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if content, ok := candidates[0].(map[string]interface{})["content"].(map[string]interface{}); ok {
			if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
				if text, ok := parts[0].(map[string]interface{})["text"].(string); ok {
					answer = text
				}
			}
		}
	}

	c.JSON(200, gin.H{"answer": answer})
}
