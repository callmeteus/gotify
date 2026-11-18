import {IVersion} from './types';

export interface IConfig {
    url: string;
    downloadBaseUrl: string;
    register: boolean;
    version: IVersion;
    oidc: boolean;
    oidcAutoRedirect: boolean;
    localAuth: boolean;
    oidcIdpName: string;
}

declare global {
    interface Window {
        config?: Partial<IConfig>;
    }
}

const config: IConfig = {
    url: 'unset',
    downloadBaseUrl: 'https://forgejo.svs.e7.dev/forgejo/gotfy/releases/download/latest/',
    register: false,
    version: {commit: 'unknown', buildDate: 'unknown', version: 'unknown'},
    oidc: false,
    oidcAutoRedirect: false,
    localAuth: true,
    oidcIdpName: 'OIDC',
    ...window.config,
};

export function set<Key extends keyof IConfig>(key: Key, value: IConfig[Key]): void {
    config[key] = value;
}

export function get<K extends keyof IConfig>(key: K): IConfig[K] {
    return config[key];
}
