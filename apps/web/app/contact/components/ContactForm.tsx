import React from "react";

const ContactForm = () => {
  return (
    <div>
      <form className="flex flex-col gap-4">
        <div>
          <label className="block text-gray-700 font-medium mb-2">
            Your Name
          </label>
          <input
            type="text"
            id="name"
            className="w-full p-4 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition duration-300"
            placeholder="Enter your name"
          />
        </div>

        <div>
          <label className="block text-gray-700 font-medium mb-2">
            Your Email
          </label>
          <input
            type="email"
            id="email"
            className="w-full p-4 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition duration-300"
            placeholder="Enter your email"
          />
        </div>

        <div>
          <label className="block text-gray-700 font-medium mb-2">
            Subject
          </label>
          <input
            type="text"
            id="subject"
            className="w-full p-4 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition duration-300"
            placeholder="What is this regarding?"
          />
        </div>

        <div>
          <label className="block text-gray-700 font-medium mb-2">
            Your Message
          </label>
          <textarea
            id="message"
            className="w-full p-4 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition duration-300"
            placeholder="Write your message here..."
          ></textarea>
        </div>

        <button
          type="submit"
          className="w-full bg-black text-white font-semibold py-4 px-6 rounded-xl text-lg transition duration-300 shadow-lg hover:shadow-xl"
        >
          Send Message
        </button>
      </form>
    </div>
  );
};

export default ContactForm;
