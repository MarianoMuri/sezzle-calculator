# AI prompts used

Tool: **Claude (Anthropic)**, used in an agentic chat session with access to a shell to write, run and test code.

## 1. Initial prompt

The assignment brief was pasted verbatim (in Spanish, as received by email), with no extra instructions:

> Desarrollar una aplicación de calculadora full-stack con un frontend en React y un microservicio backend. El frontend debe consumir la API del backend para realizar operaciones aritméticas básicas y avanzadas. Centrarse en un diseño limpio, código mantenible y arquitectura comprobable.
>
> Requisitos — Operaciones: suma, resta, multiplicación, división. Opcional: exponenciación, raíz cuadrada, porcentaje.
> Interfaz de usuario (React): interfaz intuitiva para introducir datos y mostrar resultados; validación de entrada y manejo de errores; diseño adaptable (compatibilidad básica con móviles).
> Backend (API REST): exponer endpoints para las operaciones; validar la entrada y gestionar casos excepcionales (división por cero, datos no válidos); devolver resultados en JSON.
> No funcional: código limpio, legible e idiomático; pruebas unitarias para ambas capas; documentación (configuración, uso de la API y fundamentos del diseño); opcional Dockerfile.
> Restricciones: React (se prefiere TypeScript); backend en Go.
> Entregables: repositorio Git, README con setup, ejemplos de API y decisiones de diseño, informe de pruebas y cobertura, Dockerfile opcional.
> Dedicar 2–4 horas. Priorizar corrección, claridad y mantenibilidad por encima de funciones adicionales.

## 2. How the output was used

The AI produced the initial implementation, tests and documentation from that prompt, running `go test`, `vitest`, the TypeScript compiler and a headless-browser screenshot to verify the result. I then reviewed the code, ran it locally, and own every design decision described in the README.

## 2. How the output was used

Claude (Anthropic) was used to generate the initial implementation, unit tests, and documentation.

The application was subsequently run and tested locally on macOS using Go 1.27.3 and Node.js 22.20.0.

The following checks were completed:
- Backend and frontend started successfully.
- Addition and division-by-zero error handling were manually verified.
- Backend unit tests passed with 100% statement coverage.
- All 50 frontend tests passed, with 100% statement coverage and 98.41% branch coverage.
- The frontend production build completed successfully.

ChatGPT was used to provide step-by-step guidance for local setup, execution, testing, and documentation review.


