import {Page} from 'puppeteer';
import {afterAll, beforeAll, describe, expect, it} from 'vitest';
import * as auth from './authentication';
import {GotifyTest, newTest} from './setup';
import {waitForExists} from './utils';

let page: Page;
let gotify: GotifyTest;

beforeAll(async () => {
    gotify = await newTest();
    page = gotify.page;
});

afterAll(async () => await gotify.close());

describe('Presence controls', () => {
    it('does login', async () => await auth.login(page));
    it('shows manual stage buttons on the message page', async () => {
        await waitForExists(page, '#presence-controls button', "I'm at my computer");
        const labels = await page.$$eval('#presence-controls button', (buttons) =>
            buttons.map((button) => button.textContent?.trim())
        );
        expect(labels).toContain("I'm up");
        expect(labels).toContain("I'm at my computer");
        expect(labels).toContain('Not at PC');
    });
    it('publishes a manual override', async () => {
        await page.evaluate(() => {
            const button = Array.from(document.querySelectorAll('#presence-controls button')).find(
                (element) => element.textContent?.trim() === "I'm at my computer"
            );
            (button as HTMLButtonElement).click();
        });
        await page.waitForFunction(
            () => document.querySelector('#presence-controls')?.textContent?.includes('·'),
            {timeout: 10000}
        );
        const state = await page.evaluate(async () =>
            fetch('/presence').then((response) => response.json())
        );
        expect(state[0].state).toBe('at_desk');
        expect(state[0].source).toBe('manual');
    });
});
