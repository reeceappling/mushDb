import {
    AllEntries,
    Data,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import {IsValidNote, Note} from "@/app/components/formSubcomponents/notes";
import {CheckArrayType} from "@/app/components/common";
import {ExamplePicsWithNotesIncoming} from "@/app/components/formSubcomponents/contaminations";

export interface PicWithNotesForm {
    time: number
    img: string,
    notes: AllEntries<Note>,
}

export interface NewPicWithNotesForm {
    time: number
    img: File | undefined,
    notes: AllEntries<Note>,
}

export function InitialPicsEntries(initialPics?: PicWithNotesIncoming[]): SplitAllEntries<PicWithNotesForm,NewPicWithNotesForm>{
    let initialEntries: Data<PicWithNotesForm>[] = initialPics===undefined?[]:initialPics.map((p)=>{
        return {
            data:{
                time:p.time,
                img:p.location,
                notes:{
                    existing:(p.notes || []).map((n)=>{
                        return {data: n,disabled: false}
                    }),
                    new:[],
                },
            },
            disabled: false,
        }
    })
    return {existing:initialEntries,new:[]}
}

const PicsEndpoint = "/db/images/"

export function ImageLocationFor(str: string) {
    if (str==="test.jpg"){
        return "https://picsum.photos/seed/fungitracker/300/200" // Static image from picsum (Lorem Ipsum for images) with a size of 300px tall 200px wide
    }
    return PicsEndpoint+str
}

export interface PicWithNotesIncoming {
    time: number
    location: string
    notes?: Note[]
}

export function IsValidPicWithNotesIncoming(input: any): boolean {
    return (
        typeof input === 'object' &&
        'time' in input && typeof input.time === 'number' &&
        'location' in input && typeof input.location === 'string' &&
        (('notes' in input)?Array.isArray(input.notes)&&CheckArrayType(input.notes,IsValidNote):true)
    )
}
export const ExamplePicWithNotesIncoming: PicWithNotesIncoming = ExamplePicsWithNotesIncoming[0]
