import {Note} from "@/app/components/formSubcomponents/notes";
import {
    ExamplePicWithNotesIncoming,
    PicWithNotesIncoming
} from "@/app/components/formSubcomponents/picWithNotes";
import {SelectorProps} from "@/app/components/selector";
import {EntryPerms} from "@/app/components/perms";
import {ExamplePicsWithNotesIncoming} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";

export function TestFruitOK(){
    const a: FruitData = {
        _id: "(FRUIT ID HERE)",
        creationDate: 1,
        species: "(SPECIES)",
        subspecies: "(SUBSPECIES)",
        genSpore: 7,
        transfersOut: ["(TRANSFER OUT 1)","(TRANSFER OUT 2)"],
        prints: ["(PRINT1)","(PRINT2)"],
        parentType: "bag",
        parent: "(PARENT ID)",
        pics: ExamplePicsWithNotesIncoming,
        disposed: Date.now()+5000,
        mostRecentImage: ExamplePicWithNotesIncoming,
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
}

export interface FruitData {
    _id: string
    creationDate: number
    species: string
    subspecies?: string
    genSpore?: number
    transfersOut?: string[]
    prints?: string[]
    parentType?: string
    parent?: string
    pics?: PicWithNotesIncoming[]
    disposed?: number
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

// TODO: DO WE EVEN NEED THIS?????
export function FruitSelector(sp: SelectorProps<FruitData>){
    return <div>{"Fruit selector not implemented right now"}</div>
    // TODO: ??????????????????????
    // return RecentSelector<FruitData>({
    //     msgTxt: ChannelTextNewFruit,
    //     recentEndpt: "fruits",
    //     assertType: AssertFruit,
    //     closeTxt: "Close Fruit List",
    //     //createTxt: "Create Fruit",
    //     //newForm: NewFruit, // TODO: new fruits only from fruiting chamber creation???
    //     createEndpt: "fruit", // TODO: keep????
    //     lowercase: "fruit",
    //     inline: (inlineIn: InlineProps<FruitData>)=>{return FruitInline(inlineIn)},
    // })(sp)
}

export const ChannelTextNewFruit = "newFruit"