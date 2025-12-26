"use client";

import "./welcomepage.css";
import Navbar from "./Navbar.jsx";
import Footer from "./Footer.jsx";
import HowItWorks from "./HowItWorks.jsx";

export default function WelcomePage() {
  return (
    <div className="welcome-page-wrapper">
      <Navbar />

      <div className="welcome-page-content">
        <div className="welcome-page-section" id="features">
          features
        </div>
        <HowItWorks />
      </div>
      <Footer />
    </div>
  );
}
