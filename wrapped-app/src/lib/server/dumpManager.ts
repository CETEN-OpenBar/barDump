import { exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs';
import path from 'path';
import { env } from '$env/dynamic/private';
const PYTHON_EXECUTABLE = env.PYTHON_EXECUTABLE || 'python3';
const BACKEND_DIR = env.BACKEND_DIR || 'backend/scripts';
const RAW_DATA_DIR = env.RAW_DATA_DIR || 'data/raw';
const PROCESSED_DATA_DIR = env.PROCESSED_DATA_DIR || 'data/processed';
const CONFIG_FILE = env.CONFIG_FILE || 'data/dumps_config.json';

const execPromise = promisify(exec);

// On résout les chemins par rapport à la racine du projet (un niveau au dessus de wrapped-app)
const PROJECT_ROOT = path.resolve(process.cwd(), '..');

export interface DumpConfig {
    active_dump: string;
    dumps: Record<string, DumpInfo>;
}

export interface DumpInfo {
    title: string;
    type: string;
    file: string;
    logo1?: string;
    logo2?: string;
    start_date?: string;
    end_date?: string;
}

export const dumpManager = {
    getConfig(): DumpConfig {
        const configPath = path.resolve(PROJECT_ROOT, CONFIG_FILE);
        return JSON.parse(fs.readFileSync(configPath, 'utf-8'));
    },

    saveConfig(config: DumpConfig) {
        const configPath = path.resolve(PROJECT_ROOT, CONFIG_FILE);
        fs.writeFileSync(configPath, JSON.stringify(config, null, 2));
    },

    async exportData() {
        console.log('Exporting data from MongoDB via Go API...');
        const rawDataDir = path.resolve(PROJECT_ROOT, RAW_DATA_DIR);
        
        const response = await fetch('http://localhost:8080/api/export', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ output: rawDataDir })
        });
        
        if (!response.ok) {
            throw new Error(`Export failed: ${await response.text()}`);
        }
    },

    async processStats(params: {
        id: string;
        start_date?: string;
        end_date?: string;
    }) {
        console.log(`Processing stats for ${params.id} via Go API...`);
        const inputPath = path.resolve(PROJECT_ROOT, RAW_DATA_DIR, 'transactions.json');
        const outputPath = path.resolve(PROJECT_ROOT, PROCESSED_DATA_DIR, `transactions_${params.id}.json`);
        const accountsPath = path.resolve(PROJECT_ROOT, RAW_DATA_DIR, 'accounts.json');

        const requestBody = {
            input: inputPath,
            output: outputPath,
            accounts: accountsPath,
            start_date: (params.start_date && params.id !== 'all') ? params.start_date : '',
            end_date: (params.end_date && params.id !== 'all') ? params.end_date : ''
        };

        const response = await fetch('http://localhost:8080/api/process', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(requestBody)
        });

        if (!response.ok) {
            throw new Error(`Process failed: ${await response.text()}`);
        }

        return `transactions_${params.id}.json`;
    },

    async sendEmails(dumpId: string) {
        return new Promise((resolve, reject) => {
            this.sendEmailsStream(dumpId, () => {}).then(resolve).catch(reject);
        });
    },

    async sendEmailsStream(dumpId: string, onMessage: (msg: { text: string, type: 'info'|'success'|'error' }) => void, signal?: AbortSignal) {
        console.log(`Sending emails for dump ${dumpId} via Go API...`);
        const config = this.getConfig();
        const dump = config.dumps[dumpId];
        if (!dump) throw new Error('Dump not found');

        const accountsFile = path.resolve(PROJECT_ROOT, PROCESSED_DATA_DIR, dump.file);
        
        onMessage({ text: 'Lancement de la tâche d\'envoi d\'emails (Go)...', type: 'info' });
        
        const response = await fetch('http://localhost:8080/api/send-emails', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                accounts_file: accountsFile,
                dump_id: dumpId
            }),
            signal
        });

        if (!response.ok) {
            onMessage({ text: `Erreur: ${await response.text()}`, type: 'error' });
            throw new Error(`Mail sending failed`);
        }

        if (response.body) {
            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            try {
                while (true) {
                    const { done, value } = await reader.read();
                    if (done) break;
                    
                    const chunk = decoder.decode(value, { stream: true });
                    const lines = chunk.split('\n');
                    
                    for (const line of lines) {
                        if (line.startsWith('data:')) {
                            const text = line.slice(5).trim();
                            if (!text) continue;
                            
                            let type: 'info'|'success'|'error' = 'info';
                            if (text.includes('successfully')) type = 'success';
                            if (text.includes('Failed') || text.includes('Erreur')) type = 'error';
                            if (text.includes('déjà envoyé')) type = 'info';
                            if (text === ': ping') continue; // ignore keep-alive
                            
                            onMessage({ text, type });
                        }
                    }
                }
            } catch (err: any) {
                if (err.name === 'AbortError') {
                    onMessage({ text: 'Envoi des emails annulé par l\'utilisateur.', type: 'error' });
                    throw err;
                }
                throw err;
            }
        }

        return true;
    },

    async listAvailableLogos() {
        const staticDir = path.resolve(PROJECT_ROOT, 'wrapped-app/static');
        const files = fs.readdirSync(staticDir);
        return files
            .filter(f => /\.(png|jpg|jpeg|svg|webp)$/i.test(f))
            .map(f => `/${f}`);
    },

    getMailTemplate(): string {
        const templatePath = path.resolve(PROJECT_ROOT, 'data/mail_template.html');
        if (fs.existsSync(templatePath)) {
            return fs.readFileSync(templatePath, 'utf-8');
        }
        return '';
    },

    saveMailTemplate(content: string) {
        const templatePath = path.resolve(PROJECT_ROOT, 'data/mail_template.html');
        fs.writeFileSync(templatePath, content, 'utf-8');
    },

    async saveLogo(file: File) {
        const staticDir = path.resolve(PROJECT_ROOT, 'wrapped-app/static');
        const buildClientDir = path.resolve(PROJECT_ROOT, 'wrapped-app/build/client');
        
        const buffer = Buffer.from(await file.arrayBuffer());
        const fileName = file.name.replace(/\s+/g, '_');
        
        // Write to static dir (persisted via volume)
        const filePath = path.join(staticDir, fileName);
        fs.writeFileSync(filePath, buffer);
        
        // Write to build client dir so SvelteKit Node server can serve it immediately
        if (fs.existsSync(buildClientDir)) {
            const buildPath = path.join(buildClientDir, fileName);
            fs.writeFileSync(buildPath, buffer);
        }
        
        return `/${fileName}`;
    }
};
