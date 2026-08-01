import { dumpManager } from '$lib/server/dumpManager';

export const load = async () => {
    const config = dumpManager.getConfig();
    
    return {
        activeDump: config.dumps[config.active_dump],
        activeId: config.active_dump
    };
};
