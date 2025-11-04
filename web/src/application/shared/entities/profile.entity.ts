interface ConstructorParams {
  id: string;
  username: string;
}

export class Profile {
  constructor(params: ConstructorParams) {
    this.id = params.id;
    this.username = params.username;
  }

  readonly id: string;
  readonly username: string;
}
