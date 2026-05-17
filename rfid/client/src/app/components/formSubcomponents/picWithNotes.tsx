import {
    AllEntries,
    AreaProps,
    Data,
    FormListArea,
    GroupProps,
    SplitAllEntries, SplitEntriesV2
} from "@/app/components/formSubcomponents/shared";
import NotesArea, {NotesAreaOld, IsValidNote, Note, NotesGrid} from "@/app/components/formSubcomponents/notes";
import {ChangeEvent} from "react";
import DateArea from "@/app/components/formSubcomponents/date";
import {CheckArrayType} from "@/app/components/common";
import {ExamplePicsWithNotesIncoming} from "@/app/components/formSubcomponents/contaminations";

// export interface PicWithNotesNew {
//     time: number
//     location: string,
//     notes?: Note[],
// }

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

export function InitialPicsEntries(initialPics?: PicWithNotesIncoming[]): SplitEntriesV2<PicWithNotesForm,NewPicWithNotesForm>{
    let initialEntries: Data<PicWithNotesForm>[] = initialPics===undefined?[]:initialPics.map((p)=>{
        let ns: Note[] = p.notes || []
        return {
            data:{
                time:p.time,
                img:p.location,
                notes:{existing:ns.map((n)=>{
                        return {data: n,disabled: false}
                    }),new:[]},
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