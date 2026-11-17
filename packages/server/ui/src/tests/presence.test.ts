import {Page} from 'puppeteer';
import {afterAll, beforeAll, describe, expect, it} from 'vitest';
import * as auth from './authentication';
import * as selector from './selector';
import {GotifyTest, newTest} from './setup';
import {count, innerText, waitForCount, waitForExists} from './utils';

let page: Page;
let gotify: GotifyTest;

beforeAll(async () => {
    gotify = await newTest();
    page = gotify.page;
});

afterAll(async () => await gotify.close());

const clickByText = async (text: string) =>
    page.evaluate((label) => {
        const button = Array.from(document.querySelectorAll('button')).find(
            (element) => element.textContent?.trim() === label
        ) as HTMLButtonElement | undefined;
        button?.click();
    }, text);

describe('Presence', () => {
    it('does login', async () => await auth.login(page));

    it('navigates to presence', async () => {
        await page.click('#navigate-presence');
        await waitForExists(page, selector.heading(), 'Presence');
        expect(page.url()).toContain('/presence');
    });

    it('starts with no reports', async () => {
        expect(await count(page, '#presence-table tbody tr')).toBe(0);
    });

    it('publishes a manual away report', async () => {
        await clickByText("I'm away");
        await waitForCount(page, '#presence-table tbody tr', 1);
        const state = await innerText(
            page,
            '#presence-table tbody tr:nth-child(1) td:nth-child(2)'
        );
        expect(state).toBe('Away');
        const source = await innerText(
            page,
            '#presence-table tbody tr:nth-child(1) td:nth-child(3)'
        );
        expect(source).toBe('manual');
    });

    it('overrides with at the PC', async () => {
        await clickByText("I'm at the PC");
        await page.waitForFunction(() =>
            Array.from(document.querySelectorAll('#presence-table tbody tr td:nth-child(2)')).some(
                (cell) => cell.textContent?.trim() === 'At the PC'
            )
        );
    });
});
