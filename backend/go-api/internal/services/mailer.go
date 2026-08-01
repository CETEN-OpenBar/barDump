package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"
)

func sendHTMLEmail(senderEmail, senderPassword, receiverEmail, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", senderEmail, senderPassword, "smtp.gmail.com")

	msg := []byte("To: " + receiverEmail + "\r\n" +
		"From: " + senderEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		htmlBody)


	err := smtp.SendMail("smtp.gmail.com:587", auth, senderEmail, []string{receiverEmail}, msg)
	if err != nil {
		log.Printf("Failed to send email to %s: %v", receiverEmail, err)
		return err
	}
	log.Printf("Email sent successfully to %s", receiverEmail)
	return nil
}

func ProcessAndSendEmails(ctx context.Context, accountsFile, dumpID string, logChan chan string) {
	defer close(logChan)
	sendLog := func(msg string) {
		log.Println(msg)
		if logChan != nil {
			logChan <- msg
		}
	}

	senderEmail := strings.TrimSpace(strings.Trim(os.Getenv("SMTP_EMAIL"), "\"'"))
	senderPassword := strings.TrimSpace(strings.Trim(os.Getenv("SMTP_PASSWORD"), "\"'"))
	if senderEmail == "" || senderPassword == "" {
		sendLog("Erreur: Les variables d'environnement SMTP_EMAIL et SMTP_PASSWORD doivent être définies.")
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://dump.bar.telecomnancy.net"
	}

	if dumpID == "" {
		dumpID = "all"
	}

	trackingFile := os.Getenv("TRACKING_FILE")
	if trackingFile == "" {
		trackingFile = fmt.Sprintf("../../data/email_sent_%s.json", dumpID)
	}

	// Load tracking file
	sentUsersMap := make(map[string]bool)
	if tfData, err := os.ReadFile(trackingFile); err == nil {
		var sentUsers []string
		if json.Unmarshal(tfData, &sentUsers) == nil {
			for _, u := range sentUsers {
				sentUsersMap[u] = true
			}
		}
	}

	accData, err := os.ReadFile(accountsFile)
	if err != nil {
		sendLog(fmt.Sprintf("Error: The file %s was not found.", accountsFile))
		return
	}
	var data map[string]interface{}
	if err := json.Unmarshal(accData, &data); err != nil {
		sendLog(fmt.Sprintf("Error: Could not decode the JSON from %s.", accountsFile))
		return
	}

	users, ok := data["utilisateurs"].([]interface{})
	if !ok {
		return
	}

	// Template est maintenant en dur dans data/
	templateFile := "../../data/mail_template.html"
	var templateStr string
	if t, err := os.ReadFile(templateFile); err == nil {
		templateStr = string(t)
	} else {
		sendLog(fmt.Sprintf("Erreur: Le fichier template %s est introuvable. Erreur: %v", templateFile, err))
		return
	}

	logo1Url := os.Getenv("LOGO1_URL")
	if logo1Url == "" {
		logo1Url = baseURL + "/logo.png"
	}
	logo2Url := os.Getenv("LOGO2_URL")

	var logosHtml string
	if logo1Url != "" && logo2Url != "" {
		logosHtml = fmt.Sprintf(`<div style="margin-bottom: 25px; text-align: center;">
                    <img src="%s" alt="Logo 1" width="150" style="display: inline-block; margin: 0 10px; border: 0; max-width: 45%%; height: auto; vertical-align: middle;">
                    <img src="%s" alt="Logo 2" width="150" style="display: inline-block; margin: 0 10px; border: 0; max-width: 45%%; height: auto; vertical-align: middle;">
                </div>`, logo1Url, logo2Url)
	} else {
		logosHtml = fmt.Sprintf(`<div style="margin-bottom: 25px; text-align: center;">
                    <img src="%s" alt="Logo Bar" width="150" style="display: block; margin: 0 auto; border: 0; max-width: 100%%; height: auto;">
                </div>`, logo1Url)
	}

	for _, userIntf := range users {
		select {
		case <-ctx.Done():
			sendLog("Opération annulée par l'utilisateur.")
			return
		default:
		}

		user, ok := userIntf.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := user["account_name"].(string)
		email, _ := user["email"].(string)
		accountID, _ := user["account_id"].(string)

		if name == "" || email == "" || accountID == "" {
			sendLog(fmt.Sprintf("Skipping user due to missing data: %v", user))
			continue
		}

		if sentUsersMap[accountID] {
			sendLog(fmt.Sprintf("Skipping %s (%s) - Email déjà envoyé.", email, accountID))
			continue
		}

		dumpLink := fmt.Sprintf("%s/dump/%s/%s", baseURL, dumpID, accountID)

		htmlBody := strings.ReplaceAll(templateStr, "{name}", name)
		htmlBody = strings.ReplaceAll(htmlBody, "{dump_link}", dumpLink)
		htmlBody = strings.ReplaceAll(htmlBody, "{base_url}", baseURL)
		htmlBody = strings.ReplaceAll(htmlBody, "{logos_html}", logosHtml)
		//ne retire pas l'adresse en brut
		err := sendHTMLEmail(senderEmail, senderPassword, email, "[BAR] Le BarDump est enfin arrivé !", htmlBody)
		if err == nil {
			sendLog(fmt.Sprintf("Email sent successfully to %s", email))
			sentUsersMap[accountID] = true
			var updatedSentUsers []string
			for k := range sentUsersMap {
				updatedSentUsers = append(updatedSentUsers, k)
			}
			outBytes, _ := json.Marshal(updatedSentUsers)
			os.WriteFile(trackingFile, outBytes, 0644)
		} else {
			sendLog(fmt.Sprintf("Failed to send email to %s: %v", email, err))
		}

		time.Sleep(2 * time.Second)
	}

	sendLog("Process finished")
}
