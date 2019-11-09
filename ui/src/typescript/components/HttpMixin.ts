import Vue from "vue";
import {Mixin} from 'vue-mixin-decorator';

@Mixin
export default class HttpMixin extends Vue {

  async getRequest(url: string): Promise<Response> {
    let response: Response = new Response();

    try {
      response = await fetch(url);
    } catch (error) {
      console.log(error);
    }

    return response;
  }

  async postRequest(url: string, data: {}): Promise<Response> {
    let response: Response = new Response();

    try {
      response = await fetch(url, {
        "method": "POST",
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      });
    } catch (error) {
      console.log(error);
    }

    return response;
  }

  async putRequest(url: string, data: {}): Promise<Response> {
    let response: Response = new Response();

    try {
      response = await fetch(url, {
        "method": "PUT",
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      });
    } catch (error) {
      console.log(error);
    }

    return response;
  }
};