import {ExamplePicWithNotesIncoming, PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {Note} from "@/app/components/formSubcomponents/notes";
import {ExamplePicsWithNotesIncoming, TestNotes} from "@/app/components/formSubcomponents/contaminations";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector from "@/app/components/selector";
import {SporePrintSelector} from "@/app/components/sporePrintClient";


export function TestSporePrintOk(){
    return new SporePrintData({
        _id: "(SUBSTR ID HERE)",
        parent: "(PARENT ID)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        color: "Black",
        density: "Average",
        pics: ExamplePicsWithNotesIncoming,
        sale: "SALE ID",
        disposed: Date.now(),
        mostRecentImage: ExamplePicWithNotesIncoming,
        notes: TestNotes,
        lastUpdated: 789,
        acl: TestAcl(),
    })
}

export interface SporePrintData {
    _id: string
    parent?: string // Only empty if purchased and not printed yourself
    species: string
    subspecies?: string
    creationDate: number
    color?: string
    density?: string
    pics?: PicWithNotesIncoming[]
    sale?: string
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    disposed?: number
    lastUpdated: number
    acl: ACL
}
export class SporePrintData {
    // Accept a single object containing the fields
    constructor(init?: Partial<SporePrintData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "sporePrint"
    }
}

export function SporePrintSelectorCloseable({onSelect,hideDisposed}:{onSelect: (val?: SporePrintData)=>void,hideDisposed?:boolean}) {
    const doSel = (val?: SporePrintData):void=>{
        if (!val){
            return
        }
        onSelect(val)
    }
    return <CloseableSelector<SporePrintData> props={{
        allowCreation: false, // TODO: ok?
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Spore Print List",
        //createTxt: "Create Fruit", // TODO: ok?
        lowercase: "sporePrint",
        //creatorInPage: sp.creatorInPage, // TODO: ok?
        //createEndpt: "fruit", // TODO: ok?
        createSelector:(selHdl: (onSelect: SporePrintData) => void)=>{
            return <SporePrintSelector hideDisposed={hideDisposed} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: do we want a creator anywhere for this?
        // createCreator:(selHdl: (onSelect: FruitData) => void)=>{
        //     return <FruitSelector handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}