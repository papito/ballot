import Vue from "vue";
import {Mixin} from 'vue-mixin-decorator';

@Mixin
export default class HttpMixin extends Vue {

  async getRequest(url: string): Promise<{}> {
    let json = null;

    try {
      const response = await fetch(url);
      json = await response.json();
    } catch (error) {
      console.log(error);
    }

    return json;
  }

  async postRequest(url: string, data: {}): Promise<{}> {
    let json = null;

    try {
      const response = await fetch(url, {
        "method": "POST",
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      });
      json = await response.json();
    } catch (error) {
      console.log(error);
    }

    return json;
  }

  async putRequest(url: string, data: {}): Promise<{}> {
    let json = null;

    try {
      const response = await fetch(url, {
        "method": "PUT",
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      });
      json = await response.json();
    } catch (error) {
      console.log(error);
    }

    return json;
  }
};