"use client";
import "./welcomepage.css"

export default function HowItWorks() {
    return (
        <div className="welcome-page-section how-it-works" id="how-it-works">
            <h2 className="how-title">How it works</h2>
            <p className="how-subtitle">
               Create powerful automations in three simple steps
            </p>

            <div className="how-steps">
                <div className="how-step">
                    <div className="step-circle">1</div>
                    <h3>Choose your trigger</h3>
                    <p>Select the app and event that starts your automation. For example, "When I receive an email".</p>
                    <span className="step-example">Gmail, Dropbox, Twitter, Slack…</span>
                </div>

                <div className="how-step">
                    <div className="step-circle">2</div>
                    <h3>Add an action</h3>
                    <p>Choose what happens next. You can add multiple actions and create complex workflows with conditional logic.</p>
                    <span className="step-example">Save to Drive, Send message…</span>
                </div>

                <div className="how-step">
                    <div className="step-circle">3</div>
                    <h3>Activate and relax</h3>
                    <p>Turn on your automation and let it run in the background. Monitor performance and tweak as needed.</p>
                    <span className="step-example">Running 24/7 automatically</span>
                </div>
            </div>

            <div className="how-cta">
                <h3>Ready to automate your workflow?</h3>
                <p>Join millions of users who save hours every week with automated workflows. Start for free, no credit card required.</p>
                <a className="how-cta-btn" href="/login">Get Started for Free</a>
            </div>
        </div>
    );
}
