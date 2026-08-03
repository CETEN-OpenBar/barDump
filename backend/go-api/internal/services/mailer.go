package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"
)

func sendHTMLEmail(host, port, senderEmail, senderPassword, receiverEmail, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", senderEmail, senderPassword, host)

	msg := []byte("To: " + receiverEmail + "\r\n" +
		"From: " + senderEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		htmlBody)

	addr := host + ":" + port

	if port == "465" {
		tlsconfig := &tls.Config{
			ServerName: host,
		}

		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			return err
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, host)
		if err != nil {
			return err
		}

		if err = c.Auth(auth); err != nil {
			return err
		}

		if err = c.Mail(senderEmail); err != nil {
			return err
		}

		if err = c.Rcpt(receiverEmail); err != nil {
			return err
		}

		w, err := c.Data()
		if err != nil {
			return err
		}

		_, err = w.Write(msg)
		if err != nil {
			return err
		}

		err = w.Close()
		if err != nil {
			return err
		}

		return c.Quit()
	}

	err := smtp.SendMail(addr, auth, senderEmail, []string{receiverEmail}, msg)
	if err != nil {
		log.Printf("Failed to send email to %s: %v", receiverEmail, err)
		return err
	}
	log.Printf("Email sent successfully to %s", receiverEmail)
	return nil
}

func ProcessAndSendEmails(ctx context.Context, accountsFile, dumpID string, debugMode bool, debugEmail string, logChan chan string) {
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

	smtpHost := strings.TrimSpace(strings.Trim(os.Getenv("SMTP_HOST"), "\"'"))
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}

	smtpPort := strings.TrimSpace(strings.Trim(os.Getenv("SMTP_PORT"), "\"'"))
	if smtpPort == "" {
		smtpPort = "587"
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

	type DumpConfig struct {
		Logo1 string `json:"logo1"`
		Logo2 string `json:"logo2"`
	}
	type ConfigFileStruct struct {
		Dumps map[string]DumpConfig `json:"dumps"`
	}

	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "../../data/dumps_config.json"
	}

	logo1Url := baseURL + "/logo.png"
	logo2Url := ""

	if configData, err := os.ReadFile(configFile); err == nil {
		var cfg ConfigFileStruct
		if err := json.Unmarshal(configData, &cfg); err == nil {
			if dumpData, exists := cfg.Dumps[dumpID]; exists {
				if dumpData.Logo1 != "" {
					// ensure the logo path starts with a slash
					if !strings.HasPrefix(dumpData.Logo1, "/") {
						dumpData.Logo1 = "/" + dumpData.Logo1
					}
					logo1Url = baseURL + dumpData.Logo1
				}
				if dumpData.Logo2 != "" {
					if !strings.HasPrefix(dumpData.Logo2, "/") {
						dumpData.Logo2 = "/" + dumpData.Logo2
					}
					logo2Url = baseURL + dumpData.Logo2
				}
			}
		}
	}

	// Allow overriding with env variables if explicitly set
	if envLogo1 := os.Getenv("LOGO1_URL"); envLogo1 != "" {
		logo1Url = envLogo1
	}
	if envLogo2 := os.Getenv("LOGO2_URL"); envLogo2 != "" {
		logo2Url = envLogo2
	}

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

		if sentUsersMap[accountID] && !debugMode {
			sendLog(fmt.Sprintf("Skipping %s (%s) - Email déjà envoyé.", email, accountID))
			continue
		}

		targetEmail := email
		if debugMode {
			if debugEmail == "" {
				sendLog("Erreur: debugEmail est vide mais debugMode est activé.")
				continue
			}
			targetEmail = debugEmail
			sendLog(fmt.Sprintf("[DEBUG MODE] Envoi de l'email pour %s vers %s", email, targetEmail))
		}

		dumpLink := fmt.Sprintf("%s/dump/%s/%s", baseURL, dumpID, accountID)

		htmlBody := strings.ReplaceAll(templateStr, "{name}", name)
		htmlBody = strings.ReplaceAll(htmlBody, "{dump_link}", dumpLink)
		htmlBody = strings.ReplaceAll(htmlBody, "{base_url}", baseURL)
		htmlBody = strings.ReplaceAll(htmlBody, "{logos_html}", logosHtml)
		//ne retire pas l'adresse en brut
		subject := "[BAR] Le BarDump est enfin arrivé !"
		if debugMode {
			subject = "[DEBUG] " + subject
		}
		err := sendHTMLEmail(smtpHost, smtpPort, senderEmail, senderPassword, targetEmail, subject, htmlBody)
		if err == nil {
			sendLog(fmt.Sprintf("Email sent successfully to %s", email))
			if !debugMode {
				sentUsersMap[accountID] = true
				var updatedSentUsers []string
				for k := range sentUsersMap {
					updatedSentUsers = append(updatedSentUsers, k)
				}
				outBytes, _ := json.Marshal(updatedSentUsers)
				os.WriteFile(trackingFile, outBytes, 0644)
			}
		} else {
			sendLog(fmt.Sprintf("Failed to send email to %s: %v", targetEmail, err))
		}

		time.Sleep(2 * time.Second)
	}

	sendLog("Process finished")
}
