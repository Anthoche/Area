import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './app.html',
  styleUrls: ['./app.css']
})
export class AppComponent {
  email = '';
  password = '';

  login() {
    console.log("Logging in with", this.email, this.password);
    alert(`Email: ${this.email}\nPassword: ${this.password}`);
  }
}
