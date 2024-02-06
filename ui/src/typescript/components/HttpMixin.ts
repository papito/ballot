/*
 * The MIT License
 *
 * Copyright (c) 2020,  Andrei Taranchenko
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

import Vue from 'vue';
import {Mixin} from 'vue-mixin-decorator';
import {ValidationError} from '../errors';

@Mixin
export default class HttpMixin extends Vue {

  afterCreate() {
    this.clearErrorState();
  }

  public clearErrorState() {
    // critical error
    const el: HTMLElement | null = document.getElementById(`criticalError`);
    if (!el) {
      // tslint:disable-next-line:no-console
      console.log('Could not display error: \'criticalError\' ID is missing');
      return;
    }
    el.textContent = '';
    el.style.display = 'none';

    // form errors
    let errElementList = document.querySelectorAll('.is-invalid');
    // tslint:disable-next-line:no-shadowed-variable
    errElementList.forEach(el => {
      el.classList.remove('.is-invalid');
    });
    errElementList = document.querySelectorAll('.invalid-feedback');
    // tslint:disable-next-line:no-shadowed-variable
    errElementList.forEach(el => {
      el.textContent = '';
    });
  }

  public async getRequest(url: string): Promise<Response> {
    this.clearErrorState();

    const res: Response =  await fetch(url);
    await this.processErrors(res);
    return res;
  }

  public async postRequest(url: string, data: {}): Promise<Response> {
    this.clearErrorState();

    const res: Response =  await fetch(url, {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });

    await this.processErrors(res);
    return res;
  }

  public async putRequest(url: string, data: {}): Promise<Response> {
    this.clearErrorState();

    const res: Response =  await fetch(url, {
      method: 'PUT',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });

    await this.processErrors(res);
    return res;
  }

  public async processErrors(res: Response) {
    // tslint:disable-next-line:triple-equals
    if (res.status == 400) {
      const json = await res.json();
      const fieldName = json.field;
      const errorStr = json.error;

      const targetEl: HTMLElement | null = document.getElementById(`this.${fieldName}`);
      if (!targetEl) {
        throw new Error(`Could not get element [this.${fieldName}] by ID`);
      }

      const errorStrEl: HTMLElement | null = targetEl.nextElementSibling as HTMLElement;
      if (!errorStrEl || !errorStrEl.classList.contains('invalid-feedback')) {
        throw new Error(`Could not get element [.invalid-feedback] adjacent to [this.${fieldName}]`);
      }

      targetEl.classList.add('is-invalid');
      errorStrEl.textContent = errorStr;

      throw new ValidationError(json);
    }

    // tslint:disable-next-line:triple-equals
    if (res.status == 500) {
      const json = await res.json();
      throw new Error(`Internal Server Error: ${json.message}`);
    }
  }

  public showError(err: Error) {
    if (err instanceof ValidationError) {
      // tslint:disable-next-line:no-console
      console.log('Validation error occurred: ' + JSON.stringify(err.data));
    } else {
      this.showCriticalError(err);
    }
  }

  public showCriticalError(err: Error) {
    const el: HTMLElement | null = document.getElementById(`criticalError`);

    if (!el) {
      // tslint:disable-next-line:no-console
      console.log('Could not display error: \'criticalError\' ID is missing');
      return;
    }
    el.textContent = err.message;
    el.style.display = 'block';
  }
}
