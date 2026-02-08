export class Listener {
  constructor() {
    this.handlers = [];
  }

  onChange(handler) {
    this.handlers.push(handler);
  }

  notify(changed) {
    for (const handler of this.handlers) {
      try {
        handler(changed);
      } catch (err) {
        console.error('[ConfigClient] change handler error:', err);
      }
    }
  }
}
