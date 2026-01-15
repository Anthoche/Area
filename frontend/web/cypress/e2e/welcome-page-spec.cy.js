describe('template spec', () => {
  it('passes', () => {
    cy.visit('http://localhost:5173')
  })
});

it('Navbar tests', function() {});

it('Footer links tests', function() {
  cy.visit('http://localhost:5173/')
  
  cy.get('#root ul.navbar-list.navbar-list-logo h2').click();
  cy.get('#root ul.navbar-list li:nth-child(1) a').click();
  cy.get('#root ul.navbar-list li:nth-child(2) a').click();
  cy.get('#root li.navbar-login-btn a').click();
  
});