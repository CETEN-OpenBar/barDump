import { error } from '@sveltejs/kit';
import * as fs from 'fs';
import path from 'path';
import { dumpManager } from '$lib/server/dumpManager';
import type { AllData } from '$lib/types/dump';
import { PROCESSED_DATA_DIR } from '$env/static/private';

export const load = (({ params }) => {
    const userId = params.user_id;
    const yearId = params.year;
    
    const config = dumpManager.getConfig();
    const dumpInfo = config.dumps[yearId];

    if (!dumpInfo) {
        throw error(404, "Dump not found in config");
    }

    // On résout le chemin des données traitées
    const PROJECT_ROOT = path.resolve(process.cwd(), '..');
    const dataPath = path.resolve(PROJECT_ROOT, PROCESSED_DATA_DIR, dumpInfo.file);

    if (!fs.existsSync(dataPath)) {
        console.error(`Data file not found at: ${dataPath}`);
        throw error(404, "Data file not found");
    }

    const data: AllData = JSON.parse(fs.readFileSync(dataPath, 'utf-8'));
    const user = data.utilisateurs.find(u => u.account_id === userId);

    if (!user) {
        throw error(404, 'User not found');
    }

    return {
        user,
        dumpInfo
    };
});
