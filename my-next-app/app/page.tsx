"use client";

import { useState } from 'react';
import Image from 'next/image';

export default function Home() {
  const [command, setCommand] = useState('');
  const [response, setResponse] = useState('');
  const [output, setOutput] = useState('');

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const res = await fetch('http://localhost:8080/analyze', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ command }),
    });
    const data = await res.json();
    setResponse(data.message || data.error);
    setOutput(JSON.stringify(data.output, null, 2) || '');
  };

  return (
    <div className="grid grid-cols-2 gap-4 min-h-screen p-8 pb-20 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <main className="flex flex-col gap-8 col-span-2 items-center sm:items-start">

        <ol className="list-inside list-decimal text-sm text-center sm:text-left font-[family-name:var(--font-geist-mono)]">
          <li className="mb-2">
            Get started by editing
          </li>
        </ol>

        <form onSubmit={handleSubmit} className="w-full">
          <input
            type="text"
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            placeholder="Enter command"
            className="border p-2 rounded w-full text-black"
          />
          <button type="submit" className="mt-2 p-2 bg-blue-500 text-white rounded w-full">
            Analyze
          </button>
        </form>
      </main>

      <div className="bg-black text-white p-4 rounded h-96 overflow-y-auto">
        <h2 className="text-lg font-bold">Terminal 1</h2>
        <pre>{response}</pre>
      </div>
    </div>
  );
}