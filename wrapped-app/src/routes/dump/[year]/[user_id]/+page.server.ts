import { error } from '@sveltejs/kit';
import * as fs from 'fs';
import { dumpManager } from '$lib/server/dumpManager';
import type { AllData } from '$lib/types/dump';

export const load = (({ params }) => {
    const userId = params.user_id;
    const yearId = params.year;
    
    const config = dumpManager.getConfig();
    const dumpInfo = config.dumps[yearId];

    if (!dumpInfo) {
        throw error(404, "Dump not found in config");
    }

    // On résout le chemin des données traitées
    const dataPath = dumpManager.getProcessedDataPath(dumpInfo.file);

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
