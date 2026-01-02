"use client";
import "./welcomepage.css"

export default function PopUseCases() {
    return (
      <div className="welcome-page-section popular-use-cases" id="use-cases">
          <h2>Popular use cases</h2>
          <div className="use-cases-list">
              <ul>
                  <li>Automatically save email attachments to cloud storage</li>
                  <li>Create tasks from customer support tickets</li>
                  <li>Send notifications for important events</li>
                  <li>Post new blog content to all social media channels</li>
                  <li>Sync contacts across multiple platforms</li>
                  <li>Generate reports and send to your team</li>
              </ul>
          </div>
      </div>
    )
};