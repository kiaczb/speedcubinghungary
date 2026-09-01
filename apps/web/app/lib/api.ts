/**
 * `serverApiUrl()` a Docker service nevére mutat (`http://api:8080`), amit csak a
 * konténer-hálózaton belül lehet feloldani — böngészőből ne hívd.
 * `clientApiUrl()` a böngészőből induló fetch-ekhez való; a `NEXT_PUBLIC_*`
 * változók build időben égnek bele a bundle-be, ezért módosításuk után web
 * rebuild kell.
 */

const LOCAL_API_URL = "http://localhost:8080";

export function serverApiUrl(): string {
  return (
    process.env.SERVER_API_URL || process.env.NEXT_PUBLIC_API_URL || LOCAL_API_URL
  );
}

export function clientApiUrl(): string {
  return process.env.NEXT_PUBLIC_API_URL || LOCAL_API_URL;
}
