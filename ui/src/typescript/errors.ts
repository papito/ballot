export class ValidationError extends Error {
  public data: {};

  constructor(data: {}) {
    super("Validation error");
    Object.setPrototypeOf(this, new.target.prototype);

    this.data = data;
  }
}
