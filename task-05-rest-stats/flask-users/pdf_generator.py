import os
import pika
import sys
import json
import config
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import getSampleStyleSheet
from reportlab.lib.units import cm
from reportlab.platypus import SimpleDocTemplate, Paragraph


def start_generator():
    print(' [GENERATOR] PDF generator has started')
    connection = pika.BlockingConnection(pika.ConnectionParameters(host=config.mq_address))
    channel = connection.channel()

    channel.queue_declare(queue='pdf_gen')

    def callback(ch, method, properties, body):
        print(" [GENERATOR] Received %r" % body)
        data = json.loads(body.decode('utf-8'), strict=False)
        create_pdf(data['filename'], data['text'])
        print(" [GENERATOR] Generated pdf successfully")

    channel.basic_consume(queue='pdf_gen', on_message_callback=callback, auto_ack=True)

    print(' [GENERATOR] Waiting for messages. To exit press CTRL+C')
    channel.start_consuming()


def create_pdf(filename: str, text: str):
    doc = SimpleDocTemplate(f"{filename}.pdf", pagesize=A4,
                            rightMargin=2 * cm, leftMargin=2 * cm,
                            topMargin=2 * cm, bottomMargin=2 * cm)

    doc.build([Paragraph(text.replace("\n", "<br />"), getSampleStyleSheet()['Normal']), ])


if __name__ == '__main__':
    try:
        start_generator()
    except KeyboardInterrupt:
        print('Interrupted')
        try:
            sys.exit(0)
        except SystemExit:
            os._exit(0)
