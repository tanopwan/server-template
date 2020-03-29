function addElement() {
    const rootElement = document.getElementById("root");
    const element = document.createElement("h1");
    element.textContent = "Hello, from javascript!";
    rootElement.append(element);
}
